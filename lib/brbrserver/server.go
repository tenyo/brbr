package brbrserver

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/ipsn/go-libtor"
	"github.com/tenyo/brbr/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	metagramsDir   = "metagrams"
	privateKeyFile = "ed25519_private_key"
	pidfile        = "/tmp/brbr.pid"
)

// messengerServer is used to implement pb.MessengerServer
type messengerServer struct {
	ID      string
	RecvDir string
	pb.UnimplementedMessengerServer
}

// SendMetagram handles metagram sending/receiving
func (s *messengerServer) SendMetagram(ctx context.Context, in *pb.Metagram) (*pb.Metagram, error) {
	response := "OK"

	// write to disk (organized by the from address)
	path := fmt.Sprintf("%s/%s", s.RecvDir, in.GetFrom())
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0700)
		if err != nil {
			log.Printf("failed to create output dir: %v", err)
			response = "Received but failed to save"
		}
	}
	filename := fmt.Sprintf("%s/%s", path, in.GetId())

	log.Printf("Received metagram %s from %s (size %d bytes), saving to %s", in.GetId(), in.GetFrom(), proto.Size(in), filename)

	out := fmt.Sprintf("ID: %s\nCreated_at: %s\nFrom: %s\n\n%s\n", in.GetId(), in.GetCreatedAt().AsTime().String(), in.GetFrom(), in.GetContent())
	if err := ioutil.WriteFile(filename, []byte(out), 0600); err != nil {
		log.Printf("failed to save metagram: %v", err)
		response = "Received but failed to save"
	}

	return &pb.Metagram{
		Id:        in.GetId(),
		CreatedAt: timestamppb.Now(),
		From:      s.ID,
		Content:   response,
	}, nil
}

func Start(dataDir string) {
	if dataDir == "" {
		dataDir = "."
	}

	key, err := loadKey(dataDir)
	if err != nil {
		log.Fatalf("failed to load key: %v", err)
	}

	// create metagram output dir
	metagramsPath := dataDir + "/" + metagramsDir
	if _, err := os.Stat(metagramsPath); os.IsNotExist(err) {
		log.Printf("Creating metagrams output directory %s", metagramsPath)
		if err := os.Mkdir(metagramsPath, 0700); err != nil {
			log.Fatalf("failed to create metagrams dir: %v", err)
		}
		if err := os.Mkdir(metagramsPath+"/"+"received", 0700); err != nil {
			log.Fatalf("failed to create metagrams received dir: %v", err)
		}
	}
	log.Printf("All received metagrams will be saved in %s/received", metagramsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Initializing Tor")

	// use embedded tor
	t, err := tor.Start(ctx, &tor.StartConf{
		DataDir:                dataDir + "/.tordata-server",
		ProcessCreator:         libtor.Creator,
		UseEmbeddedControlConn: true,
		RetainTempDataDir:      true,
		DebugWriter:            nil, //os.Stderr
	})
	if err != nil {
		log.Panicf("Failed to start tor: %v", err)
	}
	defer t.Close()

	// use system tor
	//t, err := tor.Start(ctx, &tor.StartConf{DebugWriter: os.Stderr})
	//t, err := tor.Start(ctx, nil)
	//if err != nil {
	//        log.Fatalf("failed to start tor: %v", err)
	//}
	//defer t.Close()

	log.Printf("Starting onion service, please wait ...")

	// Create an onion v3 service to listen on a random local port but show as 80
	onion, err := t.Listen(ctx, &tor.ListenConf{
		Key:          ed25519.PrivateKey(key),
		Version3:     true,
		RemotePorts:  []int{80},
		Detach:       true,
		DiscardKey:   true,
		NonAnonymous: false,
	})
	if err != nil {
		log.Fatalf("failed to start onion service: %v", err)
	}

	// write our address to a file so client can read it when sending
	if err := ioutil.WriteFile(dataDir+"/address", []byte(onion.ID), 0600); err != nil {
		log.Printf("failed to save address: %v", err)
	}

	log.Printf("Onion service listening at %v", onion.ID)

	grpc := grpc.NewServer()
	pb.RegisterMessengerServer(grpc, &messengerServer{
		ID:      onion.ID,
		RecvDir: metagramsPath + "/" + "received",
	})

	// setup signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChan)
	go func() {
		s := <-signalChan
		log.Printf("got %v signal, attempting graceful shutdown", s)
		cancel()
		grpc.GracefulStop()
	}()

	// start gRPC server (blocking)
	if err := grpc.Serve(onion); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}

	if err := os.Remove("address"); err != nil {
		fmt.Printf("failed to remove address file: %v\n", err)
	}
	log.Println("clean shutdown")
}

func Stop() {
	if _, err := os.Stat(pidfile); err != nil {
		fmt.Println("server not running (pid file not found)")
		return
	}

	data, err := ioutil.ReadFile(pidfile)
	if err != nil {
		fmt.Println("server not running (pid file not found)")
		return
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Printf("unable to read and parse pid in %s: %v\n", pidfile, err)
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("unable to find process ID [%v]: %v\n", pid, err)
		return
	}

	// TODO: figure out why we have to send signal twice
	err = process.Signal(syscall.SIGINT)
	time.Sleep(time.Second)
	err = process.Signal(syscall.SIGINT)
	if err != nil {
		fmt.Printf("failed to kill process ID [%v]: %v\n", pid, err)
		return
	}

	if err = os.Remove(pidfile); err != nil {
		fmt.Printf("failed to remove pid file %s: %v\n", pidfile, err)
	}

	fmt.Printf("Stopped server process [%v]\n", pid)
}

func loadKey(dataDir string) ([]byte, error) {
	generateKey := false

	log.Printf("Loading private key")

	privateKeyPath := dataDir + "/" + privateKeyFile
	f, err := os.ReadFile(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			generateKey = true
		} else {
			return nil, err
		}
	}

	if generateKey {
		log.Print("existing key not found, generating new private key")

		key, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}

		// save to file
		if err := os.WriteFile(privateKeyPath, []byte(hex.EncodeToString(key.PrivateKey())), 0600); err != nil {
			return nil, err
		}

		return key.PrivateKey(), nil
	}

	pkey, err := hex.DecodeString(strings.TrimSuffix(string(f), "\n"))
	if err != nil {
		return nil, err
	}

	return pkey, nil
}

func SavePid(pid int) error {
	if err := ioutil.WriteFile(pidfile, []byte(strconv.Itoa(pid)+"\n"), 0600); err != nil {
		return err
	}

	return nil
}

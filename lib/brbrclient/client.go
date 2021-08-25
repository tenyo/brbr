package brbrclient

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/google/uuid"
	"github.com/ipsn/go-libtor"
	"github.com/tenyo/brbr/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Send(rcptAddr string) {
	var rcptId, msg string

	if strings.HasSuffix(rcptAddr, ".onion") {
		rcptAddr = rcptAddr + ":80"
	} else {
		rcptAddr = rcptAddr + ".onion:80"
	}

	rcptId = strings.TrimSuffix(rcptAddr, ".onion:80")

	// try to determine our address
	from := "anonymous"
	if _, err := os.Stat("address"); err == nil {
		data, err := ioutil.ReadFile("address")
		if err == nil {
			from = strings.TrimSpace(string(data))
		}
	}
	fmt.Printf("Using from address %s\n", from)

	// read msg input
	var msgLines []string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("Enter message (Ctrl+D to end):")
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msgLines = append(msgLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	msg = strings.Join(msgLines, "\n")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// use embedded tor (only on linux)
	t, err := tor.Start(ctx, &tor.StartConf{
		DataDir:                "/tmp/.tordata-client",
		ProcessCreator:         libtor.Creator,
		UseEmbeddedControlConn: true,
		RetainTempDataDir:      true,
		DebugWriter:            nil,
	})
	if err != nil {
		log.Panicf("Failed to start tor: %v", err)
	}
	defer t.Close()

	// use system tor
	//t, err := tor.Start(ctx, &tor.StartConf{DebugWriter: os.Stderr})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer t.Close()

	log.Printf("Connecting to onion service %s", rcptAddr)

	dialer, err := t.Dialer(ctx, nil)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	conn, err := grpc.DialContext(ctx, rcptAddr,
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithDialer(func(rcptAddr string, timeout time.Duration) (net.Conn, error) {
			dialCtx, dialCancel := context.WithTimeout(ctx, timeout)
			defer dialCancel()
			return dialer.DialContext(dialCtx, "tcp", rcptAddr)
		}),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewMessengerClient(conn)

	id := uuid.New().String()

	log.Printf("Sending metagram %s to %s", id, rcptId)
	r, err := c.SendMetagram(ctx, &pb.Metagram{
		Id:        id,
		CreatedAt: timestamppb.Now(),
		From:      from,
		Content:   msg,
	})
	if err != nil {
		log.Fatalf("could not send metagram: %v", err)
	}
	log.Printf("Got response for metagram %s from %s: %s", id, r.GetFrom(), r.GetContent())
}

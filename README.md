[![Go Report Card](https://goreportcard.com/badge/tenyo/brbr)](https://goreportcard.com/report/tenyo/brbr)

# brbr

`brbr` is a simple command line tool for sending and receiving messages (here called metagrams) over the Tor network. When sending a metagram to another brbr instance, it establishes an anonymous encrypted connection (as v3 onion hidden service). It works behind firewall/NAT without requiring any port opening/forwarding.

Currently brbr only works on Linux - it has a built-in Tor server (thanks to https://github.com/ipsn/go-libtor) so the compiled binary has no external dependencies.

Download the latest release from https://github.com/tenyo/brbr/releases/latest

For Raspberry Pi 4 you can use the `arm64` release, and for Pi 3 (32-bit) - the `armv6` one.

## Usage

### Quickstart

From the directory where you have `brbr` run the following command to start listening (you will see the address id that other people can use to send you metagrams):
```
./brbr start
```

To send a metagram to someone you can just pipe your message to the `brbr send` command:
```
echo "do you copy?" | ./brbr send <address_id>
```

### Address id's

The address format used by brbr is the standard v3 onion address, which has 56 alpha-numeric characters and looks like this: `qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad`.
When you first start your brbr server it will automatically generate an address (which will not change later if you restart the server). Anyone who knows that address can send you a metagram, as long as your server is running.

### Output directory

By default brbr will write all of its files (including temp files, received metagrams and key) under $HOME/.brbr.
Received metagrams will be under ~/.brbr/metagrams/received

### Examples

#### Start brbr server in the background

```
$ ./brbr start
Using data dir: /home/me/.brbr
Starting in background ...
brbr started and listening in the background (pid 30029)
$ 2021/08/24 14:21:45 Loading private key
2021/08/24 14:21:45 existing key not found, generating new private key
2021/08/24 14:21:45 Creating metagrams output directory /home/me/.brbr/metagrams
2021/08/24 14:21:45 All received metagrams will be saved in /home/me/.brbr/metagrams/received
2021/08/24 14:21:45 Initializing Tor
2021/08/24 14:21:46 Starting onion service, please wait ...
2021/08/24 14:21:52 Onion service listening at qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad
```

When starting a server for the first time, it will generate a new private key and save it as `ed25519_private_key`. The public address is based on that key, so as long as you don't change or delete the private key, you will keep the same address. In our case we got `qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad`.

Any metagrams that we receive will be saved as separate files under a `metagrams/received` dir, organized by the sender address.

#### Send a metagram

To send a metagram to address `elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd` we can just do:

```
$ ./brbr send elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd
Using from address qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad
Enter message (Ctrl+D to end):
Does this even work?
^D
2021/08/24 14:27:00 Connecting to onion service elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd.onion:80
2021/08/24 14:27:10 Sending metagram 796c465f-a5c1-4175-930f-391598788570 to elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd
2021/08/24 14:27:11 Got response for metagram 796c465f-a5c1-4175-930f-391598788570 from elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd: OK
```

You will be prompted to type a message - it can be as long as you want - just press Ctrl+D on a new line when finished and it will attempt to get sent.
Above you can see that we connected successfully to the specified server and got `OK` in response.

Note that you can also pipe a message or a file into `brbr send`, so the above metagram could be done on one line:
```
echo 'Does this even work?' | ./brbr send elzpbtfjdlygwlih3ukfqlya5gfdwaok43o356lv54ug6jxa3c3mqhqd
```

At the same time on the server receiving the metagram we would see a log message like this:
```
2021/08/24 14:27:11 Received metagram 796c465f-a5c1-4175-930f-391598788570 from qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad (size 132 bytes), saving to /home/me/.brbr/metagrams/received/qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad/796c465f-a5c1-4175-930f-391598788570
```

#### Read received metagrams

Received metagrams are organized on the file system under `metagrams`/`received`/`<sender_address>`/`<metagram_id>`

The one from the above example has been saved on the receiving server in
```
metagrams/received/
  qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad/
    796c465f-a5c1-4175-930f-391598788570
```
where `qkhe5ph6hq4zrdb3fkg53jcshqk3bq2q23pgeulipgoofzsqnlediqad` is the sender address and `796c465f-a5c1-4175-930f-391598788570` is the id of the metagram.

This is what it looks like:
```
$ cat ~/.brbr/metagrams/received/qkhe6ph6hq4zrdc3fkg53jcshqk3bqbq23pgeulipgonfzsqnlediqad/796c465f-a5c1-4175-930f-391598788570
ID: 796c465f-a5c1-4175-930f-391598788570
Created_at: 2021-08-24 14:27:10.671079837 +0000 UTC
From: qkhe6ph6hq4zrdc3fkg53jcshqk3bqbq23pgeulipgonfzsqnlediqad

Does this even work?
```

Note that a sender "from" address is not required and could be just "anonymous"

#### Stop brbr server

```
$ ./brbr stop
Stopped server process [30029]
2021/08/24 14:30:08 got interrupt signal, attempting graceful shutdown
2021/08/24 14:30:08 clean shutdown
```

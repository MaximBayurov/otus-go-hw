package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &primitiveTelnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

type primitiveTelnetClient struct {
	address    string
	timeout    time.Duration
	in         io.ReadCloser
	out        io.Writer
	connection net.Conn
}

func (c *primitiveTelnetClient) Connect() error {
	dialer := &net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var err error
	c.connection, err = dialer.DialContext(ctx, "tcp", c.address)

	return err
}

func (c *primitiveTelnetClient) Close() error {
	return c.connection.Close()
}

func (c *primitiveTelnetClient) Send() error {
	scanner := bufio.NewScanner(c.in)
	for scanner.Scan() {
		_, err := c.connection.Write([]byte(scanner.Text()))
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *primitiveTelnetClient) Receive() error {
	reader := bufio.NewReader(c.connection)
	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		_, err = fmt.Fprint(c.out, data)
		if err != nil {
			return err
		}
	}
}

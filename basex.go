package basex

import (
	"bufio"
	m5 "crypto/md5"
	"fmt"
	"net"
)

type BaseXClient struct {
	*bufio.ReadWriter
	con net.Conn
}

func (b *BaseXClient) exec(cmd byte, arg string) {
	b.ReadWriter.WriteByte(cmd)
	b.send(arg)
}

func (b *BaseXClient) send(str string) {
	strLen := len(str)
	for i := 0; i < strLen; i++ {
		if str[i] == byte(00) || str[i] == byte(255) {
			b.ReadWriter.WriteString("\xFF")
		}
		b.ReadWriter.WriteByte(str[i])
	}
	b.WriteByte(byte(0))
	b.ReadWriter.Flush()
}

func (b *BaseXClient) ok() bool {
	d, _ := b.ReadWriter.ReadByte()
	return d == byte(0)
}

func New(adr string, user string, pass string) (cli *BaseXClient, err error) {
	cli = &BaseXClient{}

	cli.con, _ = net.Dial("tcp", adr)
	cli.ReadWriter = bufio.NewReadWriter(bufio.NewReader(cli.con), bufio.NewWriter(cli.con))
	ts := cli.ReadString()
	cli.send(user)
	cli.send(md5(md5(pass) + ts))

	if !cli.ok() {
		err = fmt.Errorf("Login error")
		cli = nil
	}

	return
}

func (b *BaseXClient) Close() {
	b.con.Close()
}

func (b *BaseXClient) Command(cmd string) (string, string) {
	b.send(cmd)
	result := b.ReadString()
	info := b.ReadString()
	b.ok()
	return result, info
}

func (b *BaseXClient) Query(qry string) *Query {
	b.exec(byte(0), qry)
	id := b.ReadString()
	if !b.ok() {
		panic(b.ReadString())
	}
	return &Query{id: id, cli: b, hasNext: false, lastResult: "", state: 0}
}

func (b *BaseXClient) WriteString(str string) {
	b.ReadWriter.WriteString(str)
	b.ReadWriter.WriteByte(0)
	b.ReadWriter.Flush()
}

func (b *BaseXClient) WriteByte(bte byte) {
	b.ReadWriter.WriteByte(bte)
	b.Flush()
}

func (b *BaseXClient) ReadString() (s string) {
	for {
		d, err := b.ReadWriter.ReadByte()

		if err != nil || d == byte(0) {
			break
		}

		if d == byte(255) {
			d, err = b.ReadWriter.ReadByte()
		}

		s = s + string(d)
	}
	return
}

func md5(str string) string {
	md5io := m5.New()
	md5io.Write([]byte(str))
	return fmt.Sprintf("%x", string(md5io.Sum(nil)))
}

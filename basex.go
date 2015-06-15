package basex

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
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

	var ok bool
	cli.send(user)
	if i := strings.Index(ts, ":"); i != -1 {
		ok = cli.login(user, pass, string(ts[:i]), string(ts[i+1:]))
	} else {
		ok = cli.loginLegacy(pass, ts)
	}

	if ok {
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

func (this *BaseXClient) login(user, password, realm, nonce string) bool {
	this.send(md5Hex(md5Hex(user+":"+realm+":"+password) + nonce))
	return this.ok()
}

func (this *BaseXClient) loginLegacy(password, nonce string) bool {
	this.send(md5Hex(md5Hex(password) + nonce))
	return this.ok()
}

func md5Hex(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

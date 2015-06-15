package basex

import "errors"

const (
	Q_NONE = iota
	Q_MORE
)

type Query struct {
	id         string
	lastResult string
	hasNext    bool
	cli        *BaseXClient
	state      int
}

func (q *Query) More() (result bool) {
	if q.state == Q_NONE {
		q.cli.exec(byte(4), q.id)
		q.state = Q_MORE
	}

	b, _ := q.cli.ReadByte()
	if b == byte(0) {
		if !q.cli.ok() {
			tst := q.cli.ReadString()
			panic(tst)
		} else {
			q.state = Q_NONE
		}
		q.lastResult = ""
		q.hasNext = false
	} else {
		q.lastResult = q.cli.ReadString()
		q.hasNext = true
	}
	return q.hasNext
}

func (q *Query) Next() (s string, err error) {
	if q.hasNext {
		s = q.lastResult
		err = nil
	} else {
		s = ""
		err = errors.New("Logic error")
	}
	return s, err
}

func (q *Query) Bind(name string, value string, valType string) (err error) {
	q.cli.WriteByte(byte(3))
	q.cli.send(q.id)
	q.cli.send(name)
	q.cli.send(value)
	q.cli.send(valType)
	q.cli.ReadString()
	if !q.cli.ok() {
		errTxt := q.cli.ReadString()
		err = errors.New(errTxt)
	}
	return err
}

func (q *Query) ExecToChan(c chan<- string) {
	if q.state == Q_NONE {
		q.cli.exec(byte(4), q.id)
		q.state = Q_MORE
	}

	for {
		b, _ := q.cli.ReadByte()
		if b == byte(0) {
			if !q.cli.ok() {
				tst := q.cli.ReadString()
				panic(tst)
			} else {
				q.state = Q_NONE
			}
			close(c)
			break
		} else {
			c <- q.cli.ReadString()
		}
	}
}

func (q *Query) Execute() (r string, err error) {
	q.cli.WriteByte(byte(5))
	q.cli.send(q.id)
	r = q.cli.ReadString()
	if !q.cli.ok() {
		errTxt := q.cli.ReadString()
		err = errors.New(errTxt)
	}
	return
}

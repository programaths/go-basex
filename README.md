go-basex
========

Basic BaseX Client implementation for Go language (Golang)

Example of use :

```Go
package main

import (
	//"fmt"
	"github.com/programaths/basex"
)

func main() {

	println("connecting")
	r, _ := basex.New("127.0.0.1:1984", "admin", "admin")
	//r.Command("INFO")
	// q := r.Query(`
	// 	1,2,3
	// `)

	//q.Bind("$a", "42", "xs:int")

	// for q.More() {
	// 	println("has more (no bind)")
	// 	r2s, r2e := q.Next()
	// 	println(r2s)
	// 	println(r2e)
	// }

	for i := 0; i < 3; i++ {
		q := r.Query(`
			declare variable $a external;
			declare variable $b external;
			for $i in 0 to 100
			return ($i+$a)*$b
		`)

		q.Bind("$a", "3", "xs:decimal")
		q.Bind("$b", "5", "xs:decimal")

		c := make(chan string)
		go q.ExecToChan(c)
		if c != nil {
			for a := range c {
				println(a)
			}
		}

		// for q.More() {
		// 	r2s, r2e := q.Next()
		// 	if r2e != nil {
		// 		println(r2e.Error)
		// 	} else {
		// 		println(r2s)
		// 	}
		// }
	}

	r.Close()
}
```
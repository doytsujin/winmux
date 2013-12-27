/*

    A version of win written in Go. That does terminal multiplexing.

*/

package main

import (
	"bytes"
	"code.google.com/p/goplan9/plan9/acme"
	"fmt"
	"log"
	"os"
//	"code.google.com/p/goplan9/draw"
//	"image"
	"github.com/rjkroege/winmux/ttypair"
)

func main() {
	fmt.Print("hello from winmux\n");

	log.Print("hello!");

	// take a window id from the command line
	// I suppose it could come from the environment too
	
	log.Print(os.Args[0])

	var win *acme.Win
	var err error

	// TODO(rjkroege): look up a window by name if an argument is provided
	// and connect to it.
	if len(os.Args) > 1 {
		log.Fatal("write some code to lookup window by name and connect")
	} else {
		win,err = acme.New()
	}
	if err != nil {
		log.Fatal("can't open the window? ", err.Error())
	}

	win.Fprintf("body", "hi rob")

	// You will want this to run in its own goroutine?
	/*
		Mini design note: I can imagine having a single goroutine owning the
		connection to the acme instead of using a lock. Except that the event
		reader is synchronous and it can be waiting for an arbitrary while
		on input.

		I'll stick with the lock for the moment.
	*/
	acmetowin(win)

	win.CloseFiles()
	fmt.Print("bye\n")
}

// TODO(rjkroege): move me to top
// TODO(rjkroege): figure out what I do
// TODO(rjkroege): handle what I'm suppose to do properly
type Q struct {
	p int
}


func unknown(e *acme.Event) {
			log.Printf("unknown message %c%c\n", e.C1, e.C2);
}

// Replicates the functionality of the stdinproc in win.c
// Reads the event stream from acme, updates the window and
// echos the received content
// note that the connection to the Acme is not thread safe so we 
// need a lock.
func acmetowin(win *acme.Win) {
	debug := true

	// What is this for? I'll find out.
	// I should have a separate struct tracking the Acme
	// buffer state. Thi is what is in q.

	// Note that I will need to stash the contents of the buffer in a struct
	// per pty pair.
	var q Q
	t := ttypair.New()

	for {
		if(debug) {
			log.Printf("typing[%d,%d)\n", q.p, q.p /* +ntyper */);
		}
		e, err := win.ReadEvent()
		if err != nil {
			log.Fatal("event stream stopped? ", err.Error())
		}
		if(debug) {
			log.Printf("msg %c%c q[%d,%d)... ", e.C1, e.C2, e.Q0, e.Q1);
		}
		
		// queue for lock
		// qlock(&q.lk);

		switch(e.C1) {
		default:	// be backwards compatible: ignore additional future messages.
			unknown(e)
		case 'E':	/* write to body or tag; can't affect us */
			switch(e.C2){
			case 'I', 'D':		/* body */
				if(debug) {
					log.Printf("shift typing %d... ", e.Q1 - e.Q0);
				}
				q.p += e.Q1-e.Q0;
			case 'i', 'd':		/* tag */
			default:
				unknown(e)
			}
			break;

		case 'F':	/* generated by our actions; ignore */
		case 'K', 'M':		// Keyboard or Mouse actions that edit the file
			switch(e.C2){
//			case 'I':	// text inserted into the body (This is a capital i)
//				if(e.Nr == 1 && e.Text[0] == 0x7F) {
//					// handle delete characters: delete the character.
//					// write addr, delete character
//					log.Print("ship delete off to child\n")
//					char buf[1];
//					fsprint(addrfd, "#%ud,#%ud", e.Q0, e.Q1);
//					fswrite(datafd, "", 0);
//					buf[0] = 0x7F;
//					// ship DEL off to child.
//					write(fd0, buf, 1);
//					break;
//				}
//				// We're tracking the length of the buffer in q.p
//				// this is inserting before the end.
//				if(e.Q0 < q.p){
//					if(debug)
//						log.Printf("shift typing %d... ", e.Q1-e.Q0);
//					q.p += e.Q1-e.Q0;
//				}
//				// this is typing at the end.
//				else if(e.Q0 <= q.p+ntyper){
//					if(debug)
//						log.Printf("type... ");
//					type(&e, fd0, afd, dfd);
//				}
//				break;
			case 'D':    // deleting text from the body
				n := t.Delete(e);
				// TODO(rjkroege): fold this into t? Keep buffer state
				// separate.
				q.p -= n;
				if t.Israw() && e.Q1 >= q.p+n {
					t.Sendbs(n);
				}
				break;
			case 'x':    // button 2 in the tag or body
			case 'X':
				if(e.Flag & 1 != 0 || (e.C2=='x' && e.Nr==0)){
					/* send it straight back */
					win.WriteEvent(e);
					break;
				}
				if bytes.Equal([]byte("cook"), e.Text) {
					log.Print("should set cook to 1 whatever that does.")
//					cook = 1;
					break;
				}
				if bytes.Equal([]byte("nocook"), e.Text) {
					log.Print("should clear cook")
//					cook = 0;
					break;
				}
				// Send stuff to child
				log.Printf("should send %s to child process\n", e.Text)
				// Shouldn't this also push the contents to the 
				// sendtochild(...)
			case 'l':        // button 3, tag or body
			case 'L':
				/* just send it back */
				win.WriteEvent(e);
				break;
			case 'd':        // text deleted or inserted into the tag.
			case 'i':
				break;
			default:
				unknown(e)
			}
		}
		// Release the lock.
		// qunlock(&q.lk);
	}
}
# Go concurrency expalained.

## The issue explained below with theory.

```go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

// cache is simple for this app and store the books in in-memory cache
var cache = map[int]Book{}
var randNum = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	/*
		Scenario : If the cache has the data, the data is returned from the cache
		If the database has the data, then the data is returned from the db and
		stored in the cache if it is queried for the next time.
	*/

	for i := 0; i < 10; i++ {
		id := randNum.Intn(10) + 1
		// We are creating multiple go routine ->
		// one for cache the other for database
		// meaning we are making the code to work concurrently but not
		// parallely.
		// With this the output is printed.
		go func(id int) {
			if b, ok := queryCache(id); ok {
				fmt.Println("from cache")
				fmt.Println(b)
			}
		}(id)

		go func(id int) {
			if b, ok := queryDatabase(id); ok {
				fmt.Println("from database")
				fmt.Println(b)
			}
		}(id)

		// fmt.Printf("Book not found with id: '%v'", id)

		// What happens if there is no pause in the main function?
		// We expect the data to be seen from the database query and this is
		// considered to be the side effect to be pausing the main go routine which
		// is bad. Now if we run again we see no output.

		// Why we see no output? -> We see no output because the main function does
		// not have anything to pause itself. So even thought it is generating
		// those go routines, those go routines does not have enough time for those
		// routines to complete themselves / to return. So the go programs works in
		// such a way that we will generate all these 20 routines and the exit the
		// program since there is nothing to execute.

		// So as long as we try to pause our main program with time.Sleep we will
		// be able to see the output and give the time for our go routines to complete.

		time.Sleep(150 * time.Millisecond)
	}
	// This sleep call is for the go routines to finish
	time.Sleep(2 * time.Second)
}

func queryCache(id int) (Book, bool) {
	b, ok := cache[id]
	return b, ok
}

func queryDatabase(id int) (Book, bool) {
	time.Sleep(100 * time.Millisecond)
	for _, b := range books {
		if b.ID == id {
			cache[id] = b
			return b, true
		}
	}

	return Book{}, false
}

// Challenges with Concurrency

// 1. How to run things concurrently? -> can be done using go routines
// 2. How to make our tasks to coordinate with each other? -> Coordinating taks
// Solution - WaitGroups -> they allow to coordinate tasks. We will do is make
// a go routine to wait until the rest of the other go routines are completed.
// 3. Shared Memory -> this problem can be solved using Mutexes -> allow us to
// share memory between goroutines and our application. Mutexes are going to
// allow us to protect memory that's shared between multiple goroutines to
// ensure that we have control over what's accessing that shared memory at a
// given time.
```

## The solution provided here for the first issue which is coordination of tasks using waitgroups.

```go
// code till now

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// cache is simple for this app and store the books in in-memory cache
var cache = map[int]Book{}
var randNum = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	/*
		Scenario : If the cache has the data, the data is returned from the cache
		If the database has the data, then the data is returned from the db and
		stored in the cache if it is queried for the next time.
	*/
	// We are passing the address of the wg to our functions below. You should
	// not copy a waitgroup if you are passing it around in the program, you
	// should pass pointer instead
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {

		id := randNum.Intn(10) + 1
		// We are creating multiple go routine ->
		// one for cache the other for database
		// meaning we are making the code to work concurrently but not
		// parallely.
		// With this the output is printed.

		// Since we have 2 goroutines, we need to add that to our waitgroup.
		// Everytime we are in the main function and you are about to start a
		// concurrent task or just about to kick off a Goroutine, we call an add
		// method on the waitgroup wg and add the no of task that wants to be wait
		// on. Can be done 2 ways, use Add method above the go routines you have initialized or just use it once like done below.
		// wg.Add(1)
		wg.Add(2)
		// wg *sync.WaitGroup -> pointer to the waitGroup object.
		go func(id int, wg *sync.WaitGroup) {
			if b, ok := queryCache(id); ok {
				fmt.Println("from cache")
				fmt.Println(b)
			}
			// This means that once concurrent task is completed.
			wg.Done()
		}(id, wg)
		go func(id int, wg *sync.WaitGroup) {
			if b, ok := queryDatabase(id); ok {
				fmt.Println("from database")
				fmt.Println(b)
			}
			wg.Done()
		}(id, wg)

		// fmt.Printf("Book not found with id: '%v'", id)

		// What happens if there is no pause in the main function?
		// We expect the data to be seen from the database query and this is
		// considered to be the side effect to be pausing the main go routine which
		// is bad. Now if we run again we see no output.

		// Why we see no output? -> We see no output because the main function does
		// not have anything to pause itself. So even thought it is generating
		// those go routines, those go routines does not have enough time for those
		// routines to complete themselves / to return. So the go programs works in
		// such a way that we will generate all these 20 routines and the exit the
		// program since there is nothing to execute.

		// So as long as we try to pause our main program with time.Sleep we will
		// be able to see the output and give the time for our go routines to complete.

		// time.Sleep(150 * time.Millisecond)
	}
	// This sleep call is for the go routines to finish
	// time.Sleep(2 * time.Second)

	// Wait till waitGroup counter is 0
	wg.Wait()
}

func queryCache(id int) (Book, bool) {
	b, ok := cache[id]
	return b, ok
}

func queryDatabase(id int) (Book, bool) {
	time.Sleep(100 * time.Millisecond)
	for _, b := range books {
		if b.ID == id {
			// cache[id] = b
			return b, true
		}
	}

	return Book{}, false
}

// Challenges with Concurrency

// 1. How to run things concurrently? -> can be done using go routines
// 2. How to make our tasks to coordinate with each other? -> Coordinating taks
// Solution - WaitGroups -> they allow to coordinate tasks. We will do is make
// a go routine to wait until the rest of the other go routines are completed.
// 3. Shared Memory -> this problem can be solved using Mutexes -> allow us to
// share memory between goroutines and our application. Mutexes are going to
// allow us to protect memory that's shared between multiple goroutines to
// ensure that we have control over what's accessing that shared memory at a
// given time.
```

### What's happening and why?

What's happening here is I've got a task swap that's happening with the Go
runtime. So it writes out from database, and then it jumps over to another
Goroutine that starts writing from database and prints out its book.

I've got multiple Goroutines that are accessing the Println function from
the FMT package, which is not thread safe. It's not safe to access that
concurrently. So we do have a little bit of weirdness with our output, but
that's not being caused by the WaitGroups.

We've seen that WaitGroups allow us to have our main function wait until
all of the concurrent tasks are completed before it returns. So the main
function in this example is supervising all of these other Goroutines, and
it's using WaitGroups so that it knows when they're done with their work.

```
from database
Title:          "The Gods Themselves"
Author:         "Isaac Asimov"
Published:      1973

from database
from database
from database
from database
Title:          "The Hitchhiker's Guide to the Galaxy"
Author:         "Douglas Adams"
Published:      1979

from database
Title:          "The Android's Dream"
Author:         "John Scalzi"
Published:      2006

from database
Title:          "The Hobbit"
Author:         "J.R.R. Tolkien"
Published:      1937

Title:          "The Hitchhiker's Guide to the Galaxy"
Author:         "Douglas Adams"
Published:      1979

Title:          "The Hitchhiker's Guide to the Galaxy"
Author:         "Douglas Adams"
Published:      1979

Title:          "A Tale of Two Cities"
Author:         "Charles Dickens"
Published:      1859

from database
Title:          "Les Misérables"
Author:         "Victor Hugo"
Published:      1862

from database
Title:          "A Tale of Two Cities"
Author:         "Charles Dickens"
Published:      1859

from database
Title:          "The Gods Themselves"
Author:         "Isaac Asimov"
Published:      1973
```

### In the code below we have tried to solve the problem of shared memory using mutex.

We can use RWMutex depending upon the scenario if we have asymmetrical requirement such as reading the cache more than writing.

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// cache is simple for this app and store the books in in-memory cache
var cache = map[int]Book{}
var randNum = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	/*
		Scenario : If the cache has the data, the data is returned from the cache
		If the database has the data, then the data is returned from the db and
		stored in the cache if it is queried for the next time.
	*/
	// We are passing the address of the wg to our functions below. You should
	// not copy a waitgroup if you are passing it around in the program, you
	// should pass pointer instead
	wg := &sync.WaitGroup{}

	m := &sync.Mutex{}
	for i := 0; i < 10; i++ {

		id := randNum.Intn(10) + 1
		// We are creating multiple go routine ->
		// one for cache the other for database
		// meaning we are making the code to work concurrently but not
		// parallely.
		// With this the output is printed.

		// Since we have 2 goroutines, we need to add that to our waitgroup.
		// Everytime we are in the main function and you are about to start a
		// concurrent task or just about to kick off a Goroutine, we call an add
		// method on the waitgroup wg and add the no of task that wants to be wait
		// on. Can be done 2 ways, use Add method above the go routines you have initialized or just use it once like done below.
		// wg.Add(1)
		wg.Add(2)
		// wg *sync.WaitGroup -> pointer to the waitGroup object.
		go func(id int, wg *sync.WaitGroup, m *sync.Mutex) {
			if b, ok := queryCache(id, m); ok {
				fmt.Println("from cache")
				fmt.Println(b)
			}
			// This means that once concurrent task is completed.
			wg.Done()
		}(id, wg, m)
		go func(id int, wg *sync.WaitGroup, m *sync.Mutex) {
			if b, ok := queryDatabase(id, m); ok {
				fmt.Println("from database")
				fmt.Println(b)
			}
			wg.Done()
		}(id, wg, m)

		// fmt.Printf("Book not found with id: '%v'", id)

		// What happens if there is no pause in the main function?
		// We expect the data to be seen from the database query and this is
		// considered to be the side effect to be pausing the main go routine which
		// is bad. Now if we run again we see no output.

		// Why we see no output? -> We see no output because the main function does
		// not have anything to pause itself. So even thought it is generating
		// those go routines, those go routines does not have enough time for those
		// routines to complete themselves / to return. So the go programs works in
		// such a way that we will generate all these 20 routines and the exit the
		// program since there is nothing to execute.

		// So as long as we try to pause our main program with time.Sleep we will
		// be able to see the output and give the time for our go routines to complete.

		time.Sleep(150 * time.Millisecond)
	}
	// This sleep call is for the go routines to finish
	// time.Sleep(2 * time.Second)

	// Wait till waitGroup counter is 0
	wg.Wait()
}

func queryCache(id int, m *sync.Mutex) (Book, bool) {
	// If I call Lock, then whatever called that lock, whichever goroutine locked
	// that, now owns the mutex. It's now controlling the mutex. So nothing else
	// is going to be able to access protected code until that owning goroutine
	// calls Unlock.
	m.Lock()
	b, ok := cache[id]
	m.Unlock()
	return b, ok
}

func queryDatabase(id int, m *sync.Mutex) (Book, bool) {
	time.Sleep(100 * time.Millisecond)
	for _, b := range books {
		if b.ID == id {
			m.Lock()
			cache[id] = b
			m.Unlock()
			return b, true
		}
	}

	return Book{}, false
}

// Challenges with Concurrency

// 1. How to run things concurrently? -> can be done using go routines
// 2. How to make our tasks to coordinate with each other? -> Coordinating taks
// Solution - WaitGroups -> they allow to coordinate tasks. We will do is make
// a go routine to wait until the rest of the other go routines are completed.
// 3. Shared Memory -> this problem can be solved using Mutexes -> allow us to
// share memory between goroutines and our application. Mutexes are going to
// allow us to protect memory that's shared between multiple goroutines to
// ensure that we have control over what's accessing that shared memory at a
// given time.

// MUTEX -> Mutual Exclusion lock
// a mutex can be used to protect a portion of your code so that only one task
// or only the owner of the mutex lock can access that code. So we can use
// that for to protect memory access. So we can lock the mutex, access the
// memory, and then unlock the mutex, ensuring that only one task can access
// that code at one time.

// Racing condition in our code:
// So in our code there are places where we are reading the cache at the same
// time we were trying to write the cache. line 84 b, ok := cache[id], here we
// are trying to read the data from the cache. and at line 92 cache[id] = b,
// here we are writing to cache. So line 84 is racing with line 92, we  were
// reading the cache at the same time we were trying to write the cache.

// use go run --race . -> race flag
```

# CHANNELS

## Create a channel

```go
ch := make(chan int)
```

Here, we are making a channel which is going to send and receive an
integer.

## Create a buffered channel

```go
ch := make(chan int, 5)
```

This is a channel with the internal capacity of 5 and it can save 5 messages inside the
channel without having the immediate receiver.

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	ch := make(chan int)

	wg.Add(2)
	go func(ch chan int, wg *sync.WaitGroup) {
		// Want to receive the message from the channel.
		// Since we are receiving the message from the channel the arrow is pointing away from the channel
		// This is the receiving operation.
		fmt.Println(<-ch)
		wg.Done()
	}(ch, wg)
	go func(ch chan int, wg *sync.WaitGroup) {
		// Here we are passing the message into the channel.
		// This is the sending operation.
		ch <- 42
		wg.Done()
	}(ch, wg)

	wg.Wait()
}

// Note: These two go routines don't know anything about each other.
// The second go routine knows about the channel on line 23 but it
// doesn't know how that's being used. The first go routine at line 13
// knows it's going to receive an integer from that channel but does not
// know where it came from. So these 2 go routines are completely decoupled
// from each other and the only dependency they have is the channel. Since
// it provides all the coordination between these go routines.
```

```go
package main

import (
	"fmt"
)
// Without go routines

func main() {
	ch := make(chan int)

	fmt.Println(<-ch)
	ch <- 42
}

// the problem that we have now is that channels are blocking constructs,
// especially the type of channel that we're working with right here. What
// that means is in order to receive a message from the channel, I have to
// have a message available or the application is going to block in this
// Goroutine until a message does become available. So on line 11, I'm trying
// to receive a message from the channel, but I haven't generated one yet. I
// generate one on line 12, but I'm never going to get there because I'm not
// going to get past line 10 until a message comes in.
```

```go
package main

import (
	"fmt"
)

func main() {
	ch := make(chan int)
	ch <- 42
	fmt.Println(<-ch)
}

//  If I reverse the operation, I have the opposite problem because I can't
// receive a message into the channel until there's a receiver. So in this
// case, I'm trying to push a message into the channel, but I can't because
// there's nothing listening for it. I do have something listening on line 11,
// but I can't get past line 10 because I'm blocked until I have a receiver.
// So if I try and run the application, I see that I get a deadlock condition.
```

### Error Message from above 2 program snippets.

```go
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/tmp/sandbox2270168116/prog.go:9 +0x37

[T+0000ms]
Program exited.
```

### Why we use channels with go routines?

_*Because you have to have a sender and a receiver that are available to operate with  
channels.So that's where having the Goroutines works for us because one Goroutine can  
recognize that it's looking for a message, but one isn't available, but it's okay because  
that Goroutine can just go to sleep and let another Goroutine take over. So in this case,  
if the first Goroutine fires first, it's going to try and receive a message from the  
channel. It can't, so it stops, the Go runtime will schedule the second Goroutine on,  
which is ready to send a message while the runtime recognizes, well, it can send the  
message because it's got another task that's waiting for that message to come through.*_

## Buffered Channels

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	// ch := make(chan int)
	ch := make(chan int, 1)

	wg.Add(2)
	go func(ch chan int, wg *sync.WaitGroup) {
		fmt.Println(<-ch)
		wg.Done()
	}(ch, wg)
	go func(ch chan int, wg *sync.WaitGroup) {
		ch <- 42
		ch <- 21
		wg.Done()
	}(ch, wg)

	wg.Wait()
}

// Let's try sending one more message to our channel.

// We get an deadlock condition. Why?
// Because I'm trying to send on a channel that doesn't have an active receiver.
// So once again, I'm in this deadlock condition. I can't finish out the second
// goroutine because I don't have anything to receive the message from it.

// Now let's say we need to have internal buffer within the channel.
// We can provide a second param. The internal capacity of the unbuffered channel is 0
// You have to have matched senders and receivers, otherwise you block. In this case,
// I can have one message sitting within the channel, so I don't have to have perfect
// matches between the senders and receivers.

// So in this case, if I run the application now, you see that I do get my message 42 out.
// Now I never get 27 because I put the message 27 into the channel, but I never received
// a message back out. So I have lost this message, but at least the application isn't blocking.
```

# Channel Types

```go
ch := make(chan int)

func myfunc(ch chan int) {...} // bidirectional channel

func myFunc(ch chan<- int) {...} // send-only channel
// Arrow is pointing into the channel

func myFunc(ch <-chan int) {...} // receive-only channel
// Arrow is pointing away from the channel
```

### Altering according to our need. Sending only channel and receiving only channel

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	// ch := make(chan int)
	ch := make(chan int, 1)

	wg.Add(2)
	// Receiving only
	go func(ch <-chan int, wg *sync.WaitGroup) {
		fmt.Println(<-ch)
		wg.Done()
	}(ch, wg)
	// Sending only
	go func(ch chan<- int, wg *sync.WaitGroup) {
		ch <- 42
		ch <- 21
		wg.Done()
	}(ch, wg)

	wg.Wait()
}
```

## Closing channel

```
1. Closed via the built-in close function
2. Cannot check for the closed channel.
3. Sending new message after closing the channel will create a panic
4. Receiving messages is okay
		If buffered, all buffered messages are still available
		If unbuffered, or buffer empty, you will receive zero-value
5. Use comma okay syntax to check
```

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	// ch := make(chan int)
	ch := make(chan int, 1)

	wg.Add(2)
	// Receiving only
	go func(ch <-chan int, wg *sync.WaitGroup) {
		// With this code you get the output: 0 false
		// which means that the sender channel is closed.
		// Commenting the close(ch) and adding ch <- 0
		// we see the output that the channel is not
		// closed and 0 is passed as value in the channel

		// msg, ok := <-ch

		// Using if condition to solve if the chan is open or closed in a scenario
		if msg, ok := <-ch; ok {
			// this will print only when the channel is sending the value
			fmt.Println(msg, ok)
		}

		wg.Done()
	}(ch, wg)
	// Sending only
	go func(ch chan<- int, wg *sync.WaitGroup) {
		// ch <- 0
		close(ch)
		wg.Done()
	}(ch, wg)

	wg.Wait()
}

// In this scenario we will get 0 printed as output.
// Now another scenario rises up, Did the channel receive tha value 0
// because it is closed or did the channel actually has the value 0.
```

```go 
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	// ch := make(chan int)
	ch := make(chan int, 1)

	wg.Add(2)
	// Receiving only
	go func(ch <-chan int, wg *sync.WaitGroup) {
		for msg := range ch {
			fmt.Println(msg)
		}
		wg.Done()
	}(ch, wg)
	// Sending only
	go func(ch chan<- int, wg *sync.WaitGroup) {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		// close(ch)
		wg.Done()
	}(ch, wg)

	wg.Wait()
}

// This program will work fine.
// Since we know the number of items we are providing to a channel.
// The receiver and the sender are in coordination with the number of
// value in channel.

// The challenge here is my receiving side has to know exactly how many
// messages are coming in right now. If you notice, my for loops have to
// be synchronized. Well, it's not always knowable how many messages are
// going to be coming in, so the sender might not even know how many
// messages it's going to be generating, so the receiver is even less likely
// to know how many messages are coming in.

// To address the above scenario, we can use range for loop on the receiver side.

// Now after making this changes the code will run but there will be deadlock formed.
// Output:
// ```0
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9

// [T+0000ms]

// fatal error: all goroutines are asleep - deadlock!

// goroutine 1 [semacquire]:
// sync.runtime_Semacquire(0xc000060040?)
// 	/usr/local/go-faketime/src/runtime/sema.go:62 +0x25
// sync.(*WaitGroup).Wait(0x0?)
// 	/usr/local/go-faketime/src/sync/waitgroup.go:139 +0x52
// main.main()
// 	/tmp/sandbox4129650485/prog.go:30 +0x117

// goroutine 6 [chan receive]:
// main.main.func1(0x0?, 0x0?)
// 	/tmp/sandbox4129650485/prog.go:16 +0x7d
// created by main.main
// 	/tmp/sandbox4129650485/prog.go:15 +0xae
// ```

// To solve this issue we have to close the channel on the sending side,
// it turns out, the for loop still has to have a way to know when it's done
// iterating. Well the way it knows it's done iterating is it looks for the
// channel to be closed. So if we have a close operation, then it's going
// to let the for loop know to stop iterating. So the reason that we had the
// deadlock was this goroutine up on line 17 was listening for the next message
// to come in, but nothing was generating new messages. So we were blocked in
// that goroutine, we didn't have anything telling us that we were done doing
// our work. So by closing the channel, the for loop is going to be able to detect
// that no messages are going to be coming into the channel anymore, and so it can 
// shut that goroutine down.

// PS - uncomment the close(ch) for the error to go
```
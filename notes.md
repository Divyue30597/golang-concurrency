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

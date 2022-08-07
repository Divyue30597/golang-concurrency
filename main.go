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

	m := &sync.RWMutex{}
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
		go func(id int, wg *sync.WaitGroup, m *sync.RWMutex) {
			if b, ok := queryCache(id, m); ok {
				fmt.Println("from cache")
				fmt.Println(b)
			}
			// This means that once concurrent task is completed.
			wg.Done()
		}(id, wg, m)
		go func(id int, wg *sync.WaitGroup, m *sync.RWMutex) {
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

func queryCache(id int, m *sync.RWMutex) (Book, bool) {
	// If I call Lock, then whatever called that lock, whichever goroutine locked
	// that, now owns the mutex. It's now controlling the mutex. So nothing else
	// is going to be able to access protected code until that owning goroutine
	// calls Unlock.
	m.RLock()
	b, ok := cache[id]
	m.RUnlock()
	return b, ok
}

func queryDatabase(id int, m *sync.RWMutex) (Book, bool) {
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

// use the RWMutex if the requirement is Reading the cache more
// Since in real world scenarios we will be reading more from the cache.

// we may have locked things down a little bit too much because while writing
// to the cache at the same time that I'm reading from it is bad, having
// multiple goroutines read from the cache at the same time because we're not
// mutating the memory, we're just inspecting it. So, in cases like this, we do
// have another type of mutex called an RWMutex.

// So only one task is going to be able to work within code that's guarded by
// the Lock and the Unlock, including readers. So when we're between Lock and
// Unlock, we are guaranteed that there are no readers and only one writer to
// the shared memory. Now when I want to inform the mutex that I want to read
// from the shared memory, then I can use the RLock and RUnlock methods. So,
// that's going to allow multiple readers to acquire that read lock, but when
// something stops on line 50 and it's trying to write, then the mutex is going
// to clear out all of those readers, let them finish their operations, then
// it's going to let the writer come in, make its update, and when the Unlock
// method is called then it's going to open the mutex up, and then it's going
// to allow multiple readers to access the protected memory again.


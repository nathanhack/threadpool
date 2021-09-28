# threadpool

threadpool - Manages a fixed pool of threads and fixed number of jobs.

## Usage

```
import "github.com/nathanhack/threadpool"
```

Create a pool to use X number of concurrent threads to do Y number of jobs.

```
X := -1 // -1 means use runtime.NumCPU() 
Y := 100 
pool := threadpool.New(context.Backgroud(), X, Y)
```

Then add some work. This method will block until a free thread can run the work before returning.

```
for i := 0; i < 50; i++ {
    pool.Add(func() {
        //do some work
    })
}
```

or add some work but don't wait until there is a thread to work on it.

```
for i := 0; i < 50; i++ {
    pool.AddNoWait(func() {
        //do some work
    })
}
```

Last we wait until all the threads complete.

```
pool.Wait()
```

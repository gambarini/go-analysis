# Go code analysis
This is an analysis for the code in pkg problem.

Below are the discussion about the issues found in the code with solutions.

The solution pkg has the code with the full solution

## Overview
The code looks like an attempt to concurrently calculate the sum of
a billion numbers (starting from 1) that goes throught the
following stages:

- Stage 1: adding 99 if the number is divisible by 2
- Stage 2: multiplying by 1, 2 or 3 when the division mod is 0, 1 or 2 respectivly
- Stage 3: adds the number to the total


## First Stage

The first section of the code appears to try to produce the numbers and apply the first stage calculation
```
for i := 1; i <= 1000000000; i++ {

    go func() {

        if i%2 == 0 {
            i += 99
        }

        jobs <- i

    }()
}

close(jobs)

```

#### Issues

- 1: the for iteration is spawning a billion go routines. That is not a good approach
to concurrency and will cause more harm than actually running the code faster

- 2: the for variable 'i' is not passed as a value to the go routines. It means that the value of 'i'
read by the routines will be the value it hold in the for iteration whenever the go routines are executed.
Instead of the expected value of 'i' when the routine was created. Also 'i' is changed by the routine affecting the
'i' count in the for loop. That creates an almost random 'i' value change based on when routines execute (which is umpredictable).

- 3: the jobs channel will block the routine when 'i' is set in it. That happens for all billion routines, and they will override the last
value set on it. So only the last routine that sets the value before the channel is read will sent anything to the reader.

- 4: The close of jobs channel happens right after the for loop. That means that it will very likely close the channel before a routine
created in the loop can set a value in the channel. When a routine attemps to set the closed channel it will cause a panic.


#### Solution

Considering that what we want is a producer of numbers (jobs).
```
func main() {

    go Stage1(jobs)

}

func Stage1(jobs chan int) {

    for i := 1; i <= 1000000000; i++ {

        if i%2 == 0 {
            i += 99
        }
        jobs <- i
    }

    close(jobs)
}

```

First it creates one go routine for the function Stage 1. That routine will
iterate to a billion creating the numbers and applying the first stage calculation.

Each number is set to the jobs channel. It will block the routine until the number is
read by another routine stage.

When all numbers are generated the channel is closed to indicate that nothing else
will be set in it, and the reader of the jobs channel can unblock
and continue with its normal execution.


## Second Stage

The second section appears to be an attempt to create workers to execute the second
stage calculations. It trys to consume the jobs from the previous stage, and produce results to the
next stage.
```
func worker(id int, jobs <-chan int, results chan<- int) {

    for j := range jobs {

        go func() {
            switch j % 3 {
            case 0:
                j = j * 1
            case 1:
                j = j * 2
                results <- j * 2
            case 2:
                results <- j * 3
                j = j * 3
            }
        }()
    }
}

func main() {

    jobs2 := []int{}

    for w := 1; w < 1000; w++ {
        jobs2 = append(jobs2, w)
    }

    for i, w := range jobs2 {
        go worker(w, jobs, results)
        i = i + 1
    }

    close(results)
}
```

#### Issues

- 1: the jobs2 array is controling the number of workers it wants to spawn (about 1000). Again, not a good approach
 for the same reason in the issue 1 in stage 1.

- 2: The routine worker will spawn more routines that actually calculate the result, one for each job from the previous stage.
That is unnecessary and will add to all the other routines already created.

- 3: In the worker function, 'j' the value in the jobs channel, is not passed to the new routine as a value. That will cause multiple
access to the loop variable causing a similar problem to issue 2 in stage 1.

- 4: the results channel is closed after the workers that set it are created. That will most likely cause a panic when the routine
try to set a value to the channel. Just like in issue 4 of Stage 1.


#### Solution

Considering we want a stage that calculate the second stage number, based on
the first stage numbers, and then send the result to the next stage.

```
func main() {

    go Stage2(jobs, results)
}

func Stage2(jobs chan int, results chan int) {

    for job := range jobs {
        switch job % 3 {
        case 0:
            results <- job * 1
        case 1:
            results <- job * 2
        case 2:
            results <- job * 3
        }
    }

    close(results)
}

```

Now we start a go routine that iterate on the jobs channel, and sets the results channel
with the calculated number for the next stage.

The jobs loop will block the routine until the next value is set by the previous stage.

Also the routine will block after setting a value in the results channel, until it's read by the next stage.

Once the jobs channel is closed, the loop will end and the results channel will be closed to
notify the results channel reader that nothing else will be added to it.


## Third Stage

The third section appears to try to read all results and print the sum of all the numbers.

```
    var sum int32 = 0

    for r := range results {
        sum += int32(r)
    }

    fmt.Println(sum)
```

#### Issues

- 1: The sum is made in the main function, not in a go routine. It can complicate the concurrency management
with the other routines/channels since it can run before the other stages are done.


#### Solution

Considering we need to sum all numbers in results channel, running concurrently with the previous stages routines.

```
func main() {

    sum := make(chan int)

    go Stage3(results, sum)

    fmt.Println(<-sum)

}

func Stage3(results chan int, sum chan int) {

    var total int

    for result := range results {
        total += result
    }

    sum <- total
}

```

Here the total calculation is turned into a go routine that consumes values in the results channel, then adds
to a total variable for each result received.

Once the results channel is closed (by the previous stage) the loop ends, and the sum is set in the new sum channel.

Now it can be read by the last command in the main function, that was blocked awaiting for the value, and finally printed out.

That ends the execution.

The full solution will execute the calculation in a pipeline of 3 routines
executing concurrently.


## Performance improvements

The jobs and results channel (stage 1 and 2 values), can be turned into a channel buffer and increase the speed of
execution of the routines.

It has the advantage of preventing block of routines after each value is set in a channel,
giving some room for all routines to execute for longer before blocking, or in the best case never blocking.

```
func main() {

    jobs := make(chan int, 100000)
    results := make(chan int, 100000)


```

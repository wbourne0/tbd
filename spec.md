




# statements


start -> equivalent of `go` in golang.

Comments: standard `//` and `/* */`


to start:

copy go

```
// initialization - statement to be executed prior to the loop's execution.
// condition - after each iteration, is evaluated with the results determining whether the loop should continue or not.
// (accepts any truthy value)
// mutation - after each time the block is executed, mutation is evaluated.
for [initialization ;] condition [; mutation] { block }

// same thing as go's if/else if/else
if [initialization ;] condition { block }
else if condition { block }
else { block }


Maps of string to 


```



```



struct myStruct {
    field1 string
    field2 int 
}

myStruct#m str() string {
    return m.field1
}

myStruct incr() int {
    return ++m.field1
}

///



blah 

anyways, try/catch is nice since you can throw an error and only callers that handle it will catch it, which avoids the
mess go kinda has with `a, _ = doSmthn()` by not just ignoring the error.

does this make sense ? 


a, catch err = doSmthn()
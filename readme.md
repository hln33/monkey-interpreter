# The Monkey Programming Language
An interpreter written in Go, following the book "Writing an Interpreter in Go" by Thorsten Ball. I wrote this project as a fun little exercise to learn the Go programming language.

## How to Run

### Running the REPL
```
$ go run main.go
```

### Running Tests
```
# Run all tests
$ go test ./...

# Run test in specific directory
$ go test ./parser
```

## Language Features
* Operators
  * Arithmetic (+, -, *, /)
  * Comparison (<, >, ==, !=)
  * Negation (!)
* Literals
  * Integer
  * String
  * Boolean
  * Array
  * Hashmap
  * Function
  * Null
* Statements
  * Let Statement (for defining variables)
  * Return
  * Block (for defining function or conditional bodies)
* Expressions
  * Function calls
  * Array indexing
  * Hashmap indexing
  * If conditionals
* Variables
* Closures & Higher Order Functions

## Interpreter Steps
```
     Raw Text Input
          |
          ▼
    Scanner/Lexxer
          |
          ▼
        Tokens
          |
          ▼
        Parser
          |
          ▼
  Abstract Syntax Tree
          |
          ▼
     Interpreter
          |
          ▼
     Code Executed
```

## Next Steps
* Additional Language Features
  * Operators like <=, >=
  * For/While Loops
* Supporting being able to interpret source code files (Right now the only way to interact with the interpreter is through the REPL)
* Going through the follow-up book "Writing a Compiler in Go" (in progress)

## See Also
Have a look at [Rlox](https://github.com/hln33/rlox/tree/main). A similar project I did, but following a different book called "Crafting Interpreters" and in the Rust programming language.

## Bookmark
Writing a compiler in Go - Hash - page 134
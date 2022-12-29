


cool features i've thought of: 
- For scaleable / incremental compiles, it should compile each module to a binary 
containing only the used call signatures, this way we only have to re-compile when a 
call signature changes or the function's definition changes.

- Reflection is done purely at compile time.  Anything not relevant to a call signature will be 
omitted from the resulting binary.

- Functions should purely be for a user to make code more readable; they can and should be inlined when at all possible.,
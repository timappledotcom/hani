# Syntax Highlighting Test

This file demonstrates the **syntax highlighting** capabilities of *Hani*.

## Code Blocks

Here's a Go code block:

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    name := "World"
    fmt.Printf("Hello, %s!\n", name)
    
    if len(os.Args) > 1 {
        name = os.Args[1]
    }
    
    for i := 0; i < 3; i++ {
        fmt.Println("Count:", i)
    }
}
```

And here's a JavaScript example:

```javascript
function greet(name = "World") {
    console.log(`Hello, ${name}!`);
    
    const numbers = [1, 2, 3, 4, 5];
    const doubled = numbers.map(n => n * 2);
    
    return {
        message: `Hello, ${name}!`,
        numbers: doubled
    };
}

greet("Hani");
```

## Markdown Elements

- **Bold text** and *italic text*
- `inline code` snippets
- [Links to external sites](https://example.com)
- Numbered lists:
  1. First item
  2. Second item
  3. Third item

> This is a blockquote with some important information.
> It can span multiple lines.

### Headers at different levels

#### Level 4 header

Normal paragraph text continues here.

## Python Code

```python
def factorial(n):
    """Calculate factorial of a number."""
    if n <= 1:
        return 1
    return n * factorial(n - 1)

# Example usage
result = factorial(5)
print(f"5! = {result}")

# List comprehension
squares = [x**2 for x in range(10)]
print("Squares:", squares)
```

That's all folks!

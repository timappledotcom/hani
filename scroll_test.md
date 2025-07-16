# Long Document for Scroll Testing

This document is designed to test the preview scrolling functionality in Hani.

## Section 1: Introduction

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

## Section 2: Features

- **Scrolling**: Use j/k or arrow keys to scroll
- **Navigation**: Use g to go to top, G to go to bottom
- **Smooth**: Scrolling should be smooth and responsive

## Section 3: Code Examples

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Hani!")
    fmt.Println("Testing preview scrolling...")

    for i := 0; i < 10; i++ {
        fmt.Printf("Line %d\n", i)
    }
}
```

## Section 4: More Content

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

### Subsection 4.1

Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.

### Subsection 4.2

Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.

## Section 5: Lists

1. First item
2. Second item
3. Third item
   - Nested item A
   - Nested item B
   - Nested item C
4. Fourth item
5. Fifth item

## Section 6: Tables

| Feature | Status | Notes |
|---------|--------|-------|
| Editor | ✅ Complete | Vim-like bindings |
| Preview | ✅ Complete | Live rendering |
| Scrolling | ✅ Fixed | Now works properly |
| Syntax | ✅ Complete | Markdown + code |

## Section 7: Final Content

This is the end of the test document. If you can scroll to see this content in the preview pane using j/k keys, then the scrolling fix is working correctly!

### Test Instructions

1. Open this file in Hani: `./hani scroll_test.md`
2. Switch to Preview tab with Tab
3. Try scrolling with:
   - `j` or `Down` to scroll down
   - `k` or `Up` to scroll up
   - `g` to go to top
   - `G` to go to bottom

If you can see all sections by scrolling, the fix is successful! 🎉
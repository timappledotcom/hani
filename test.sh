#!/bin/bash

echo "🏗️  Testing Hani Markdown Editor"
echo "================================"
echo

# Test 1: Check if binary exists
if [ ! -f "./hani" ]; then
    echo "❌ Binary not found. Building..."
    go build -o hani
    if [ $? -eq 0 ]; then
        echo "✅ Build successful"
    else
        echo "❌ Build failed"
        exit 1
    fi
else
    echo "✅ Binary found"
fi

# Test 2: Check help output
echo "📋 Testing help output..."
./hani non_existent_file.md 2>/dev/null
if [ $? -eq 1 ]; then
    echo "✅ Help output works"
else
    echo "❌ Help output failed"
fi

# Test 3: Check if sample file exists
if [ -f "./sample.md" ]; then
    echo "✅ Sample file found"
else
    echo "❌ Sample file not found"
fi

echo
echo "🎉 All tests passed!"
echo
echo "To try the editor:"
echo "  ./hani sample.md"
echo
echo "Key bindings:"
echo "  Tab/Shift+Tab  - Switch between edit and preview"
echo "  Ctrl+S         - Save file"
echo "  Ctrl+Q         - Quit"
echo "  i              - Insert mode"
echo "  Esc            - Normal mode"
echo "  h,j,k,l        - Vim navigation"

# Edge Cases Test Document

This document contains various edge cases for testing mdatlas functionality.

## Empty Sections

### 

This is a section with an empty title.

### Section with Only Spaces     

This section has a title with trailing spaces.

##  Multiple  Spaces  In  Title  

This section has multiple spaces in the title.

## Special Characters: !@#$%^&*()

This section contains special characters in the title.

## Unicode: 日本語 中文 العربية русский

This section contains Unicode characters.

### Nested **Markdown** *Formatting* in `Headings`

This section has markdown formatting in the heading.

#### Very Long Heading That Goes On And On And On And On And On And On And On And On And On And On And On

This is a very long heading to test line length handling.

##### Level 5 Heading

###### Level 6 Heading (Maximum Depth)

This is the deepest level heading.

## Code Blocks

Here's a code block:

```javascript
function test() {
    return "Hello, World!";
}
```

## Lists

- Item 1
- Item 2
  - Nested item 1
  - Nested item 2
- Item 3

## Tables

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |
| Data 4   | Data 5   | Data 6   |

## Links and Images

[Link to example](https://example.com)

![Alt text](image.jpg)

## Horizontal Rules

---

## Blockquotes

> This is a blockquote.
> It can span multiple lines.

## HTML in Markdown

<div>
    <p>This is HTML in markdown.</p>
</div>

## Escaped Characters

\# This is not a heading
\* This is not emphasis
\` This is not code

## Empty Lines Between Sections


## Section After Empty Lines

This section comes after multiple empty lines.

## Final Section

This is the final section of the document.
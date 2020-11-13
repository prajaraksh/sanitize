# sanitize
sanitize file name

## Installation

```sh
go get github.com/prajaraksh/sanitize
```

## Usage

```go
str := "   ..  ..  ..CON .. .. "

s := sanitize.New()
nv := s.Name(str)  // "   ..  ..  ..CON"
cv := s.Clean(str) // " . . .CON"
```

## Sanitize - CMD

### Installation

```sh
go get github.com/prajaraksh/sanitize/cmd/sanitize
```

### Usage

```sh
sanitize # print intended file name changes
sanitize -r # print intended file name changes recursively

sanitize -c # change intended file names
sanitize -c -r # change intended file names recursively
```

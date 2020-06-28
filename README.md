# oak
**A command-line desktop calculator for Unix**

oak exists because

- `dc` is too cryptic, and doesn't use readline
- `bc` and similar programs use infix notation, not RPN
- `apl` is all that, and requires a special keyboard too

oak borrows a little from all these, and Forth as well.

## Installation
TBD

## Usage
oak simulates a classic stack-based RPN calculator [RPN -- reverse Polish notation -- is also known as postfix notation]. Numbers are pushed onto a stack, and operators take the top item (or top two items) and leave a result on top of the stack.

For example

	$ oak
	> 2 3 +
	1: 5
	> 2 pi 3 sqr * *
	2: 56.54867

where the second example is `2 * pi * r^2` for `r=3`: the area of a circle with radius 3.

oak uses readline in interactive mode, allowing prior input lines to be recalled and edited.

Variables in the form `$n` allow re-use of prior results (`$4` recalls the fourth result in the session). 

A special variable `$0` acts as the "last x" operator, recalling the top-of-stack value from the last operation. For example

	> 4 sqrt
	1: 2
	> $0 +   `where $0 in this case is 4
	2: 6

Note that the result of the previous calculation remains on top of the stack, e.g.,

	> 2 1 +
	1: 3
	> 1 +
	2: 4

### The stack
It is common to label the four topmost stack items using the letters x, y, z, and w (where x resides on top of the stack). This helps explain how operators are taken from the stack and results pushed back. 

Note that in the examples below, the use of these four named slots does not indicate the stack is limited to four items; it is actually unlimited.

### Numbers
All numbers are currently evaluated as `float64` values in base 10. oak allows input in normal or scientific formats, e.g.

	1
	1.
	.1
	-1
	-0.1
	1.1e+3
	1e-3

A number is always pushed onto the top of the stack.

### Modes
By default, trigonometry functions evaluate their arguments in degrees; the mode may be changed to radians (see "mode" below).

By default, the calculator operates in base-10 floating point mode, but may be changed to an integer mode (see "base" below). 

Changing the base to binary, octal, or hexadecimal has these effects:

- input numbers are taken to be integers
- the output of integers is formatted in the correct base; e.g. with a `0x` prefix for hexadecimal numbers

If the base was changed by a conversion command ("bin", "oct", or "hex"):

- the top of stack will be converted to an integer (truncated) when the base is changed to binary/octal/hex
- other numbers (deeper in the stack) remain as floating point numbers unless disturbed, and will retain their full values if the mode is changed back

All math is integer math while the base is not decimal, and so any operation involving a floating point number may cause it to be truncated.

Truncated numbers are not restored when switching back to decimal.

For example

	$ oak
	> 2.3 8 base
	1: 2.3
	> dec
	2: 2.3
	> 3+
	3: 5.3
	> 8 base
	4: 5.3
	> 2+
	5: 007
	> dec
	6: 7

versus

	$ oak
	> 7 3.3 hex
	1: 0x0003
	> +
	2: 0x000a
	> dec
	3: 10

Binary numbers display in multiples of 8 bits, octal in multiples of 3 digits, and hexadecimal in multiples of 4. Thus 12 will show in 8 bits, but 257 will show in 16 bits in binary mode; both will show using 4 digits in hexadecimal mode, while 65536 will show using 8 hex digits.

There will be no support for converting floating point numbers into their equivalent unsigned integer form and vice versa (i.e., for debugging IEEE formats).

TODO: allow input using binary/octal/hexadecimal formats, e.g. with a `0b` or `0x` prefix. Note also that bitwise operators are not yet supported.

### Commands

oak offers the following operators:

	+      {y,x} -> x = y+x
	-      {y,x} -> x = y-x
	*      {y,x} -> x = y*x
	/      {y,x} -> x = y/x
	%      {y,x} -> x = y mod x
	**     {y,x} -> x = y to the power x

along with the following unary functions, which replace the top of stack with a new value

	abs    absolute value
	chs    change sign
	cbrt   cube root
	ceil   ceiling
	cos    cosine
	cube   cube (x ** 3)
	deg    convert radians to degrees (and change the mode)
	exp    e ** x
	fact   factorial [using gamma(x+1)]
	floor  floor
	frac   return the fractional part of the number
	ln     natural log
	log    log in base 10
	pow    10 ** x
	rad    convert degrees to radians (and change the mode)
	recp   reciprocal [1/x]
	sin    sine
	sqr    x ** 2
	sqrt   square root (x ** 1/2)
	tan    tangent
	trunc  truncate

and these binary functions

	dist   {y,x} -> x = sqrt(x**2 + y**2)
	dperc  {y,x} -> x = (x-y)/y * 100    [percent change from y to x]
	max    {y,x} -> x = max(x,y)
	min    {y,x} -> x = min(x,y)
	perc   {y,x} -> x = y*x / 100        [x percent of y]
	
as well as these operations on the stack / machine

	clr    reset top of stack to 0
	depth  push the existing stack depth onto it
	       {w,z,y,x} -> {z,y,x,#}
	drop   pop the top of stack
	       {w,z,y,x} -> {w,z,y}
	dup    duplicate the top of stack
	       {w,z,y,x} -> {z,y,x,x}
	dup2   duplicate the top two stack items in order
	       {w,z,y,x} -> {y,x,y,x}
	eng    pop the top of stack and set engineering notation
	       (scientific notation, but exponents are multiples of 3)
	fix    pop the top of stack and set fixed precision
	roll   roll the top of stack to the bottom
	       {w,z,y,x} -> {x,w,z,y}
	sci    pop the top of stack and set scientific format
	show   causes the top of stack to be the result
	swap   swap the top two items
	       {w,z,y,x} -> {w,z,x,y}

and these mode/conversion operations

	base   pop the top of stack and set base {2,8,10,16}
	       (default 10)
	mode   pop the top of stack and set the trigonometry mode
	       {"deg","rad"} (default degrees)
	
	bin    convert to integer mode, base 2
	oct    convert to integer mode, base 2
	hex    convert to integer mode, base 2
	dec    convert to normal (floating point) mode, base 10

and these constants

	e      base of natural logarithms, 2.71828
	pi     ratio of diameter to circumference, 3.14159
	phi    the "golden" ratio, 1.61803

There is also a single punctuation mark, where the comma (`,`) is used to separate lines of input (e.g., when using the `-e` option, below).

In non-interactive input, the backtick (`` ` ``) is used to start a comment that extends to the end of the line.

### Variables
At present, only "result" variables are supported, i.e. variables in the form `$1`.

A variable name in the input causes its value to be pushed onto the stack.

Result variables are automatically defined as results are printed (that is, line by line).

TODO: allow storing values into user-defined variables of the form `$name` using new operators to store (`!`) and recall (`@`).

### User-defined functions (words)
TODO: allow the creation of user-defined words (a la Forth), for example

	: name op op ... ;

where the name may then be used as a function operating against the stack. Note that there is no declaration of parameter numbers or types.

### Functions on strings
TODO

### Vector operations
TODO

## Command-line options
oak has only a few options

	-e <input>  read input from the command line
	-f <file>   read input from a file
	-v          show how the line parses for debugging

	-fix <num>  set fixed precison to <num> digits (e.g., %.3f)
	-sci <num>  set scientific format to <num> digits (e.g., %.3e) 

For example,

	$oak -e '1 2 +, 3+'
	1: 3
	2: 6
	$

If neither `-e` nor `-f` is present (the former takes precedence), oak starts an interactive REPL. Use ctrl-D to exit.

By default, oak uses Go's default floating point representation.

## To do
Here are a few of the possible enhancements:

- add a few missing functions (e.g. acos, tanh)
- bitwise operators, similar to the HP 16c
- interest-rate calculations, similar to the HP 12c
- statistical functions, similar to the HP 11c or 15c
- string functions (really?)
- vector operations
- user-defined variables
- user-defined words (a la Forth), along with logic & iteration
- oh, and we need a circular slide rule mode of operation, too ;-)

## Bugs
Here are a couple of issues relating to handling comments/whitespace/newlines:

- two commas in a row cause the remainder of input to be lost
- comments in files read with `-f` aren't quite right, either

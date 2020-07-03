# oak
**A command-line desktop calculator for Unix**

oak exists because

- `dc` is too cryptic and doesn't use readline
- `bc` and similar programs use infix notation, not RPN
- `apl` is all that, and requires a special keyboard too

oak borrows a little from all these and Forth as well.

## Installation
TBD

## Usage
oak simulates a classic stack-based RPN calculator [RPN -- reverse Polish notation -- is also known as postfix notation]. 
Numbers are pushed onto a stack, and operators take the top item (or top two items) and leave a result on top of the stack.

For example

	$ oak
	> 2 3 +
	1: 5
	> 2 pi 3 sqr * *
	2: 56.54867

where the second example is `2 * pi * r^2` for `r=3`: the area of a circle with radius 3.

oak uses readline in interactive mode, allowing prior input lines to be recalled and edited.

Variables in the form `$n` allow re-use of prior results (`$4` recalls the fourth result in the session). 

A special variable `$0` acts as the "last x" operator, recalling the top-of-stack value from the last operation. 
For example

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
It is common to label the four topmost stack items using the letters x, y, z, and w (where x resides on top of the stack). 
This helps explain how operators are taken from the stack and results pushed back. 

Note that in the examples below, the use of these four named slots does not indicate the stack is limited to four items; 
it is actually unlimited.

### Numbers
Decimal numbers (when the base is 10, which is the default) are evaluated as 64-bit floating point numbers, e.g.

	1
	1.
	.1
	-1
	-0.1
	1.1e+3
	1e-3

When the base is set to an integer mode (binary, octal, hexadecimal) numbers are evaluated as unsigned integers (up to 64 bits) and may be written with 0[bB] or 0[xX] prefixes if desired, e.g.

	0b10010001
	0177
	0x283e
	101

Note that integers without a leading 0 will be interpreted as base 10 integer values. See more below.

A number is always pushed onto the top of the stack immediately.

### Strings
A few commands take a string argument (e.g., mode), entered with double qoutes; these values are immediately pushed onto the stack. For example,

	> "rad" mode
	1: <nil>
	> 0.524 sin
	2: 0.500

TODO - consider operations on strings as data

### Display
There are three explicit display modes for floating-point values:

- fixed point
- scientific notation
- engineering notation (scientific, but exponents are always multple of 3)

These can be set from the command line of by commands (see below).
oak uses Go's default floating point representation if no display mode is set.

For example:

	> 3 recp
	1: 0.3333333333333333
	> 3 fix
	2: 0.333
	> 3 sci
	3: 3.333e-01
	> 3 eng
	4: 0.333e+00
	> 10 *
	5: 3.333e+00
	> 100 *
	6: 0.333e+03

### Modes

#### Angular mode
By default, trigonometry functions evaluate their arguments in degrees; the mode may be changed to radians (see "mode" and the degree/radians conversion operators below). For example,

	> 30 sin
	1: 0.500
	> 30 rad
	2: 0.524
	> sin
	3: 0.500

#### Base (radix)
By default, the calculator operates in base-10 floating point mode, but may be changed to an integer mode (see "base" below). 

Changing the base to binary, octal, or hexadecimal has these effects:

- input numbers are taken to be unsigned integers, with these options:
    - a `0[bB]` prefix indicates binary
    - a `0[xX]` prefix indicates hexadecimal
    - numbers with a leading 0 will be taken as octal (e.g., 0177 is decimal 127)
- the output of integers is formatted in the correct base; e.g. with a `0x` prefix for hexadecimal numbers

If the base was changed by a conversion command ("bin", "oct", or "hex"):

- the top of stack will be converted to an unsigned integer (truncated) when the base is changed to binary/octal/hex
- other numbers (deeper in the stack) remain as floating point numbers unless disturbed, and will retain their full values
  if the mode is changed back

For example

    $ oak
    > oct 127   `could have been "127 oct" also
    1: 0177
    > 234
    2: 0352
    > +
    3: 0551
    > dec
    4: 361

All math is integer math while the base is not decimal, and so any operation involving a floating point number may cause 
it to be truncated.

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

Binary numbers display in multiples of 8 bits, octal in multiples of 3 digits, and hexadecimal in multiples of 4. Thus 
12 will show in 8 bits, but 257 will show in 16 bits in binary mode; both will show using 4 digits in hexadecimal mode, 
while 65536 will show using 8 hex digits.

There will be no support for converting floating point numbers into their equivalent unsigned integer form and vice 
versa (i.e., for debugging IEEE formats).

### Variables
Variables have two forms

- "Result" variables in the form $1, $2, etc., auto-generated by evaluation (each name is the number of a result line)

	For example
	
		> 1 2 +
		1: 3
		> $1
		2: 3
		> +
		3: 6
		> 1 $3 +
		4: 7

	where `$0` is a special case representing the "last x" value.

	A result variable name in the input causes its value to be pushed onto the stack. Result variables are automatically defined as results are printed (that is, line by line) and cannot be modified by a store operation.	
		
- User-defined variables with alphanumeric names (not starting with a digit), for example `$name`

	User-defined variables must be created by a store (`!`) operation and deferenced by a recall (`@`) operation; their values are **not** immediately pushed as in the case of result variables.

	For example,
	
		> 1 2 +
		1: 3
		> $name !
		2: <nil>
		> 2 $name @+
		3: 5
		> $name@+
		4: 8

## Operations

oak offers the following floating-point binary operators:

	+      {y,x} -> x = y+x
	-      {y,x} -> x = y-x
	*      {y,x} -> x = y*x
	/      {y,x} -> x = y/x
	%      {y,x} -> x = y mod x
	**     {y,x} -> x = y to the power x

and these bitwise operations for unsigned integers:

	&      {y,x} -> x = y&x     [bitwise and]
	|      {y,x} -> x = y|x     [bitwise or]
	^      {y,x} -> x = y^x     [bitwise xor]
	<<     {y,x} -> x = y<<x    [left shift]
	>>     {y,x} -> x = y>>x    [logical right shift]
	>>>    {y,x} -> x = y>>>x   [arithmetic right shift]

	~      {x}   -> x = !x      [bitwise not]

along with the following floating-point unary functions, which replace the top of stack with a new value

	abs    absolute value
	chs    change sign
	cbrt   cube root (x ** 1/3)
	ceil   ceiling
	cos    cosine
	cube   cube (x ** 3)
	exp    e ** x
	fact   factorial [using gamma(x+1)]
	floor  floor
	frac   return the fractional part of the number
	ln     natural log
	log    log in base 10
	pow    10 ** x
	recp   reciprocal [1/x]
	sin    sine
	sqr    square (x ** 2)
	sqrt   square root (x ** 1/2)
	tan    tangent
	trunc  truncate

and these floating-point binary functions

	dist   {y,x} -> x = sqrt(x**2 + y**2)
	dperc  {y,x} -> x = (x-y)/y * 100    [percent change from y to x]
	max    {y,x} -> x = max(x,y)
	min    {y,x} -> x = min(x,y)
	perc   {y,x} -> x = y*x / 100        [x percent of y]

and these bitwise unary functions

	maskl  {x}   -> x = ^0 << (64-x), ^0 if x > 64  [left mask]
	maskr  {x}   -> x = ^0 >> (64-x), ^0 if x > 64  [right mask]
	popcnt {x}   -> x = population count of x (# of 1 bits)

and these unary functions on user variables (e.g., `$a`)

	!      {y,x} -> {}, vars[x]=y        [store]
	@      {x}   -> x = vars[x]          [recall]

as well as these operations on the stack / machine

	clr    reset top of stack to 0
	clrall reset the entire stack to empty
	depth  push the existing stack depth onto it
	       {w,z,y,x} -> {z,y,x,#}
	dump   display the stack & variables, leave stack unchanged
	       (very primitive debugging tool ;-)
	drop   pop the top of stack
	       {w,z,y,x} -> {w,z,y}
	dup    duplicate the top of stack
	       {w,z,y,x} -> {z,y,x,x}
	dup2   duplicate the top two stack items in order
	       {w,z,y,x} -> {y,x,y,x}
	eng    pop the top of stack and set engineering notation
	       (scientific notation, but exponents are multiples of 3)
	fix    pop the top of stack and set fixed precision
	load   pop a string off the stack and read the machine's
	       state from that file; overwrites the current state
	over   duplicate the second-from-top item onto the stack
	       {w,z,y,x} -> {z,y,x,y}
	roll   roll the top of stack to the bottom
	       {w,z,y,x} -> {x,w,z,y}
	save   pop a string off the stack and save the machine's
	       state into that file (for use with load or -i option)
	sci    pop the top of stack and set scientific format
	status display current modes; leaves stack unchanged
	swap   swap the top two items
	       {w,z,y,x} -> {w,z,x,y}
	top    causes the top of stack to be the result
	       (a blank line does the same thing)

and these mode/conversion operations

	mode   pop the top of stack and set the trigonometry mode
	       {"deg","rad"} (default degrees)

	deg    convert radians to degrees (and change the mode)
	rad    convert degrees to radians (and change the mode)

	base   pop the top of stack and set base {2,8,10,16}
	       (default 10)

	bin    convert to integer, set base 2
	oct    convert to integer, set base 8
	hex    convert to integer, set base 16
	dec    convert to normal (floating point) mode, base 10

and finally these constants

	e      base of natural logarithms, 2.71828
	pi     ratio of diameter to circumference, 3.14159
	phi    the "golden" ratio, 1.61803

There is also a single punctuation mark, where the comma (`,`) is used to separate lines of input (e.g., when using the
 `-e` option, below).

The backtick (`` ` ``) is used to start a comment that extends to the end of the line. (TBD: maybe use the single quote `'`, and allow backticks to mark a raw string.)

## Saved state
If you save the state of the machine with "save", that state includes

- the stack
- the "last x" value
- all user-defined variables, but not result variables
- all user-defined words (when we have that capability)

Loading state with "load" overwrites all existing machine state except result variables.

## User-defined functions (words)
TODO: allow the creation of user-defined words (a la Forth), for example

	: name op op ... ;

where the name may then be used as a function operating against the stack. 
Note that there is no declaration of parameter numbers or types.

Also, it will not be possible to allow result vars (`$1`, etc.) to be
used in words; we'll need to store the elements as tokens to allow
the state to be written out / loaded back in.

## Functions on strings
TODO

## Vector operations
TODO

## Command-line options
oak has only a few options

	-e <input>  read input from the command line
	-f <file>   read input from a file
	
	-i <file>   load a stored machine image from a file
	            (this creates a non-empty initial state)

	-fix <num>  set fixed precison to <num> digits (e.g., %.3f)
	-sci <num>  set scientific format to <num> digits (e.g., %.3e)
	-eng <num>  like scientific format, but exponents are multiples
	            of 3 only

	-rad        start in radians mode for trigonometry
	-debug      show how the line parses for debugging

For example,

	$ oak -e '1 2 +, 3+'
	1: 3
	2: 6
	$ oak -e '127 bin, oct, hex'
	1: 0b01111111
	2: 0177
	3: 0x007f

If neither `-e` nor `-f` is present (the former takes precedence), oak starts an interactive REPL. Exit with "bye" or type ctrl-D; the latter will not save any state.

## History
The REPL stores up to 50 lines of command history in `$HOME/.oakhist` which is available to your next session (through the normal operations at the prompt, e.g., up-arrow).

## Startup configuration
The machine will read the file `$HOME/.oak.yml` if it is present. The file may have both options and commands. For example,

	options:
	  trig_mode: "rad"
	  display_mode: fix
	  digits: 3
	commands:
	  - status 

then the REPL will display the new status before the first input line:

	$ oak
	base: 10 mode: rad display: fix/3
	>

The possible options are

	trig_mode        "deg" or "rad"
	display_mode     "free", "fix", "sci", "eng"
	base             10, 2, 8, 16
	digits           2, 0+

where the first value is the default in each case.

The list of commands is joined with "," and processed as a unit.
Thus

	commands:
	  - 1 2
	  - 3+
	  - sqr

is equivalent to `1 2, 3+, sqr`.

Note that the commands in the `.oak.yml` file do not leave result variables or a "last x" value when the machine starts. There will be no output unless an error occurs, in which case the machine will print the error and quit.

## To do
Here are a few possible enhancements:

- add a few missing trig functions (e.g. acos, tanh)
- vector operations
- string functions (really?)
- statistical functions, similar to the HP 11c
- interest-rate calculations, similar to the HP 12c
- user-defined words (a la Forth), along with logic & iteration
- oh, and we need a circular slide rule mode of operation, too ;-)

## Bugs
There are no open issues

# Overview

This is a golang library for reading/writing the .ini format file. The description on .ini file can be found at https://en.wikipedia.org/wiki/INI_file

# Supported .ini format

A .ini file contains one or more sections and each section contains one or more key/value pair. Following is an example of .ini file

```ini
# this is a comment line
; this is also a comment line

[section1]

key1 = value1

[section2]

key2 = value2
```

## Comments
### Comments line

A comments line is started with char '#' or ';' and it will be ignored when processing the .ini file.

```ini

# this is a comment line
; this is also a comment line

```

### inline comments

A comment can be appended in a tail of line. The inline comments must be started with ';' or '#' and its previous char must be a space.

```ini
[section1]
key1 = value1 ;this is a inline comment
key2 = value2;this is not a inline comment
```

## Multiline value

if a value is multiple line value, the value can be put between """ and """, an example:

```ini

[section1]

multi-line-key = """this is a multi-line example,
multiple line can be put in a value,
this is multiple line is just for test"""

single-line-key = this is a normal value
```

## Continuation line

If a line is too long, user can devide one line to multiple line and on the end of line the char '\\' should be put:

```ini
[section1]
key1 = this line is too long, \
we need to write it to multiple line, \
but actually it is one line from the point of user

```

## Escape char

This library supports the escape char, the escape char is started with char \\ 

|Common escape sequences Sequence |	Meaning                                             |
|---------------------------------|-----------------------------------------------------|
|\\\\ 	                          |\ (a single backslash, escaping the escape character)|
|\0                               |Null character                                       |
|\a 	                            |Bell/Alert/Audible                                   |
|\b 	                            |Backspace, Bell character for some applications      |
|\t 	                            |Tab character                                        |
|\r 	                            |Carriage return                                      |
|\n 	                            |Line feed                                            |
|\\; 	                            |Semicolon                                            |
|\\# 	                            |Number sign                                          |
|\\= 	                            |Equals sign                                          |
|\\: 	                            |Colon                                                |
|\\x???? 	                        |Unicode character with hexadecimal code point        |


## Environemnt variable support

Environment variable can be embeded in the value of the key and the environment variable will be replaced. For example:

```ini
[section1]
key1 = this value has env ${HOME}
key2 = this value has env with default ${SOME_ENV:-test},hihi
```

In the above example, the environment variable HOME is in the value of key1. So if the value of environment variable HOME is "/home/test", the value of key1 is "this value has env /home/test".

For the key2, the environemnt SOME_ENV is included and if the environment variable SOME_ENV does not exist, its value will be "test" otherwise it will be the value of SOME_ENV environment variable.

# API

## import the library

The go-ini library should be imported before using this library:

```go
import (
  ini "github.com/ochinchina/go-ini"
)
```
## Load .ini file

.ini format file or string can be loaded by the method:

### Load from a file

```go
//Load the .ini from a file
ini := ini.Load( "fileName" )

```

### Load from a string or byte array in .ini format

```go
ini_str := `[section1]
key1 = value1
key2 = value 2
`

ini := ini.Load( ini_str )
//load from a byte array

ini = ini.Load( []byte(ini_str) )

```

### Load from a io.Reader

```go

var reader io.Reader = ...

ini := ini.Load( reader )

```

### Load .ini from multiple source

The Load() method can load .ini from multiple mixed sources.

``` go
//load multiple sources: fileName, string, reader and byte array in one statement

ini := ini.Load( "fileName", ini_str, reader )
```

### Load the .ini in Ini object

The Ini class also provide a method named Load(), this method can be called multiple times and the later loaded .ini will be appended to the Ini object.

```go
//first load the .ini from a file
ini := ini.Load( "fileName" )

//append the .ini from string to the ini object
ini_str := `[section1]
key1 = value1
key2 = value 2
`
ini.Load( ini_str )

//append the .ini from a reader to the ini object
var reader io.Reader = ...
ini.Load( reader )

```

## Access the value of key in the .ini file

After loading the .ini from a file/string/reader, we can access a keya under a section. This library provides three level API to access the value of a key in a section.

### Access the value of key in Ini class level

The value of key can be accessed in Ini class level.

```go
ini := ini.Load(...)

value, err := ini.GetValue( "section1", "key1")

// if err is nil, the value is ok
if err == nil {
  //the value exists and DO something according to the value
}
```

Sometimes we need to provide a default value if the key in the section does not exist, at this time the user can provide a default value by GetValueWithDefault() method.

```go
ini := ini.Load(...)

//if the section1 or key1 does not exist, return a default value(empty string)
value := ini.GetValueWithDefault( "section1", "key1", "" )
```
### Access the value of key in Section class level

Call the GetSection() method by the section name on the Ini object at frist, and then call GetValue() on the section to get the value of key.

```go
ini := ini.Load(...)

section, err := ini.GetSection( "section1" )

if err == nil {
  value, err := section.GetValue( "key1" )
  if err == nil {
    //the value of key1 exists
  }
}
```

The method GetValueWithDefault() ask user provide a default value if the key under section does not exist, the user provided default value will be returned.

```go
ini := ini.Load(...)

section, err := ini.GetSection( "section1" )

if err == nil {
  //get the value of key1 and if the key1 does not exists, return the default empty string
  value := section.GetValueWithDefault("key1", "" )
}
```

### Access the value of key in Key class level

The value of a key can be acccessed in the Key class level also. The method Key() on the section with keyname can be called even if the key does not exist. After getting a Key object, user can call Value() method to get the value of key.
```go
ini := ini.Load(...)

section, err := ini.GetSection( "section1" )
if err == nil {
  //the Key() method always returns a Key object even if the key does not exist
  value, err := section.Key( "key1" ).Value()
  if err == nul {
    //the value in key1 exists
  }
}
```
User can provide a default value to method ValueWithDefault() on the Key object to get the value of key and if the key does not exist the default value will be returned.


```go
ini := ini.Load(...)

section, err := ini.GetSection( "section1" )
if err == nil {
  //the Key() method always returns a Key object even if the key does not exist
  value:= section.Key( "key1" ).ValueWithDefault("")
}
```

## Convert the string value to desired types

Except for getting a string value of a key, you can also ask the library convert the string to one of following types:

- bool
- int
- int64
- uint64
- float32
- float64

For each data type, this library provides two methods GetXXX() and GetXXXWithDefault() on the Ini&Section class level where the XXX stands for the Bool, Int, Int64, Uint64, Float32, Float64.

An example to ask the library convert the key to a int data type in Ini level:

```go

ini := ini.Load(...)

value, err := ini.GetInt( "section1", "key1" )

if err == nil {
  //at this time, the value of key1 exists and can be converted to integer
}

value = ini.GetIntWithDefault( "section1", "key1", 0 )

```

An example to ask the library convert the key to a int data type in Section level:
```go

ini := ini.Load(...)

section, err := ini.GetSection( "section1" )

if err == nil {
  value, err = section.GetInt( "key1" )
  if err == nil {
    //at this time the key1 exists and its value can be converted to int
  }
  
  value = section.GetIntWithDefault("key1", 0 )
}
```

An example to ask the library convert the key to a int data type in Key level:
```go

ini := ini.Load(...)
section, err := ini.GetSection( "section1" )
if err == nil {
  value, err := section.Key( "key1" ).Int()
  if err == nil {
    //at this time the key1 exists and its value can be converted to int
  }
  
  //get with default value
  value = section.Key( "key1" ).IntWithDefault( 0 )
}
```

## Add the key&value to .ini file

This library also provides API to add key&value to the .ini file.

```go

ini := ini.NewIni()

section := ini.NewSection( "section1" )
section.Add( "key1", "value1" )
```

## Save the .ini to the file

User can call the Write() method on Ini object to write the .ini contents to a io.Writer

```go

ini := ini.NewIni()
section := ini.NewSection( "section1" )
section.Add( "key1", "value1" )

buf := bytes.NewBufferString("")
ini.Write( buf )
```

If want to write to the file, there is a convinent API WriteToFile() with filename on the Ini object to write the .ini content to the file.


```go

ini := ini.NewIni()
section := ini.NewSection( "section1" )
section.Add( "key1", "value1" )

ini.WriteToFile( "test.ini" )

```

/*
A golang implemented library to read/write .ini format files.

With this library, you can load the .ini file a string, a byte array, a file and a io.Reader.

    import (
        ini "github.com/ochinchina/go-ini"
    )


    func main() {
        //load from .ini file
        ini := ini.Load( "myfile.ini")
        //load from .ini format string
        str_data := "[section1]\nkey1=value1\n[section2]\nkey2=value2"
        ini = ini.Load( str_data )

        //load .ini format byte array
        ini = ini.Load( []byte(str_data) )

        //load from io.Reader
        var reader io.Reader = ...

        ini = ini.Load( reader )

        //load from multiple source in one Load method
        ini = ini.Load( "myfile.ini", reader, str_data, bytes_data )
    }

The loaded Ini includes sections, you can access section:

    //get all the sections in the .ini
    var sections []*Section = ini.Sections()

    //get a section by Name
    var section *Section = ini.GetSection( sectionName )


Then the key in a section can be accessed by method GetXXX() and GetXXXWithDefault(defValue):
    //get the value of key
    value, err := section.GetValue( "key1")
    value = section.GetValueWithDefault("key1", "")

    //get value of key as int
    i, err := section.GetInt( "key2" )
    i = section.GetIntWithDefault( "key2" )

*/
package ini

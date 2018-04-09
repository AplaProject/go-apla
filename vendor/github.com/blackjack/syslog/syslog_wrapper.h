#ifndef SYSLOG_WRAPPER_H
#define SYSLOG_WRAPPER_H

#include <syslog.h>
#include <string.h>

//cgo currently cannot handle variable arguments list, so we have to use a wrapper
void go_syslog(int facility, const char* msg) {
    syslog(facility,"%s",msg);
}



//As said in openlog man:
//========================================================
//Please note that the string pointer ident will be retained internally by the Syslog routines.
//You must not free the memory that ident points to.
//========================================================
//
//Because of that we store ident string in static variable and overwrite it on every
//Openlog call
void go_openlog(const char* ident, int priority, int options) {

#define max_length 1001 //extra char for null termination
    static char _current_ident[max_length] = { 0 };

    //Copy ident string to static array. Extra 1 substraction is for those
    //cases when size of ident is larger than max_length-1. In that case
    //string would be unterminated, but having extra zero in the end
    //guarantees string termination.
    strncpy(_current_ident,ident,max_length-1);
    openlog(_current_ident,priority,options);
}

#endif

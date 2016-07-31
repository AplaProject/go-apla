Roadmap to v1.0 :
================
Please note Complete Support *may* not be toggled until Thrust core is stable.

- [ ] Add kiosk support (core.0.7.3)
- [X] Queue Requests prior to Object being created to matain a synchronous looking API without the need for alot of state checks
- [ ] Remove overuse of pointers in structs where modification will not take place
- [ ] Add Window Support
  - [X] Basic Support
  - [X] Refactor Connection usage
  - [ ] Complete Support 
    - Accessors (core.0.7.3)
    - Events (core.0.7.3)

- [ ] Add Menu Support
  - [X] Basic Support
  - [X] Refactor Connection usage
  - [ ] Complete Support

- [ ] Add Session Support
  - [X] Basic Support
  - [ ] Complete Support

- [X] Implement Package Connection

- [x] Seperate out in to packages other than main
  - [X] Package Window
  - [X] Package Menu
  - [X] Package Commands
  - [X] Package Spawn

- [ ] Remove func Main as this is a Library
  - [ ] Should use Tests instead

- [X] Refactor how Dispatching occurs
  - [X] We should not have to manually dispatch, there should be a registration method 

- [ ] Refactor menu.SetChecked to accept a nillable menu item pointer, so we dont have to waste resources finding the item in the Tree

- [X] Refactor CallWhen* methods, Due to the nature of using GoRoutines, there is the chance that calls will execute out of the original order they were intended.

- [X] Create a script to autodownload binaries

- [X] Refactor Logging

- [ ] SubMenus need order preservation

- [X] vendor folders need versioning

- [X] Need to fix Pathing for autoinstall and autorun. Relative paths will not work for most use cases.
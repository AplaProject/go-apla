#include <stdio.h>
#include <windows.h>
#include <shellapi.h>

#define WM_MYMESSAGE (WM_USER + 1)

#define MAX_LOADSTRING 100

HINSTANCE hInst;
TCHAR szTitle[MAX_LOADSTRING];
TCHAR szWindowClass[MAX_LOADSTRING];
wchar_t *titleWide;
const char *url;
NOTIFYICONDATA nid;

ATOM                MyRegisterClass(HINSTANCE hInstance);
HWND                InitInstance(HINSTANCE, int);
LRESULT CALLBACK    WndProc(HWND, UINT, WPARAM, LPARAM);

void set_url(const char* theUrl)
{
    url = theUrl;
}

void native_loop(const char *title, unsigned char *imageData, unsigned int imageDataLen)
{
    HWND hWnd;
    HINSTANCE hInstance = GetModuleHandle(NULL);

    MSG msg;

    titleWide = (wchar_t*)calloc(strlen(title) + 1, sizeof(wchar_t));
    mbstowcs(titleWide, title, strlen(title));

    wcscpy((wchar_t*)szTitle, titleWide);
    wcscpy((wchar_t*)szWindowClass, (wchar_t*)TEXT("MyClass"));
    MyRegisterClass(hInstance);

    hWnd = InitInstance(hInstance, FALSE); // Don't show window
    if (!hWnd)
    {
        return;
    }

    // Let's load up the tray icon
    HICON hIcon;
    {
        // This is really hacky, but LoadImage won't let me load an image from memory.
        // So we have to write out a temporary file, load it from there, then delete the file.

        // From http://msdn.microsoft.com/en-us/library/windows/desktop/aa363875.aspx
        TCHAR szTempFileName[MAX_PATH+1];
        TCHAR lpTempPathBuffer[MAX_PATH+1];
        int dwRetVal = GetTempPath(MAX_PATH+1,        // length of the buffer
                                   lpTempPathBuffer); // buffer for path
        if (dwRetVal > MAX_PATH+1 || (dwRetVal == 0))
        {
            return; // Failure
        }

        //  Generates a temporary file name.
        int uRetVal = GetTempFileName(lpTempPathBuffer, // directory for tmp files
                                      TEXT("_tmpicon"), // temp file name prefix
                                      0,                // create unique name
                                      szTempFileName);  // buffer for name
        if (uRetVal == 0)
        {
            return; // Failure
        }

        // Dump the icon to the temp file
        FILE* fIcon = fopen(szTempFileName, "wb");
        fwrite(imageData, 1, imageDataLen, fIcon);
        fclose(fIcon);
        fIcon = NULL;

        // Load the image from the file
        hIcon = LoadImage(NULL, szTempFileName, IMAGE_ICON, 64, 64, LR_LOADFROMFILE);

        // Delete the temp file
        remove(szTempFileName);
    }

    nid.cbSize = sizeof(NOTIFYICONDATA);
    nid.hWnd = hWnd;
    nid.uID = 100;
    nid.uCallbackMessage = WM_MYMESSAGE;
    nid.hIcon = hIcon;

    strcpy(nid.szTip, title); // MinGW seems to use ANSI
    nid.uFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP;

    Shell_NotifyIcon(NIM_ADD, &nid);

    // Main message loop:
    while (GetMessage(&msg, NULL, 0, 0))
    {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}


ATOM MyRegisterClass(HINSTANCE hInstance)
{
    WNDCLASSEX wcex;

    wcex.cbSize = sizeof(WNDCLASSEX);

    wcex.style          = CS_HREDRAW | CS_VREDRAW;
    wcex.lpfnWndProc    = WndProc;
    wcex.cbClsExtra     = 0;
    wcex.cbWndExtra     = 0;
    wcex.hInstance      = hInstance;
    wcex.hIcon          = LoadIcon(NULL, IDI_APPLICATION);
    wcex.hCursor        = LoadCursor(NULL, IDC_ARROW);
    wcex.hbrBackground  = (HBRUSH)(COLOR_WINDOW+1);
    wcex.lpszMenuName   = 0;
    wcex.lpszClassName  = szWindowClass;
    wcex.hIconSm        = LoadIcon(NULL, IDI_APPLICATION);

    return RegisterClassEx(&wcex);
}

HWND InitInstance(HINSTANCE hInstance, int nCmdShow)
{
    HWND hWnd;

    hInst = hInstance;

    hWnd = CreateWindow(szWindowClass, szTitle, WS_OVERLAPPEDWINDOW,
                        CW_USEDEFAULT, 0, CW_USEDEFAULT, 0, NULL, NULL, hInstance, NULL);

    if (!hWnd)
    {
        return 0;
    }

    ShowWindow(hWnd, nCmdShow);
    UpdateWindow(hWnd);

    return hWnd;
}

#define CMD_OPEN_IN_BROWSER 1001
#define CMD_COPY_LINK 1002
#define CMD_EXIT 1003

void ShowMenu(HWND hWnd)
{
    HMENU hSubMenu = CreatePopupMenu();
    POINT p;
    GetCursorPos(&p);

    hSubMenu = CreatePopupMenu();
    AppendMenuW(hSubMenu, MF_STRING | MF_GRAYED, CMD_OPEN_IN_BROWSER, titleWide);
    if (url)
    {
        AppendMenuW(hSubMenu, MF_SEPARATOR, 0, NULL);
        AppendMenuW(hSubMenu, MF_STRING, CMD_OPEN_IN_BROWSER, L"Open");
    }
    AppendMenuW(hSubMenu, MF_SEPARATOR, 0, NULL);
    AppendMenuW(hSubMenu, MF_STRING, CMD_EXIT, L"Exit");

    SetForegroundWindow(hWnd); // Win32 bug work-around
    TrackPopupMenu(hSubMenu, TPM_BOTTOMALIGN | TPM_LEFTALIGN, p.x, p.y, 0, hWnd, NULL);

}

void Exit()
{
    Shell_NotifyIcon(NIM_DELETE, &nid);
    PostQuitMessage(0);
}

void OpenInBrowser()
{
    ShellExecuteA(NULL, "open", url, NULL, NULL, SW_SHOWDEFAULT);
}

LRESULT CALLBACK WndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam)
{
    switch (message)
    {
        case WM_COMMAND:
            switch (LOWORD(wParam))
            {
                case CMD_EXIT:
                    Exit();
                    break;
                case CMD_OPEN_IN_BROWSER:
                    OpenInBrowser();
                    break;
                case CMD_COPY_LINK:
                {
                    const size_t len = strlen(url) + 1;
                    HGLOBAL hMem =  GlobalAlloc(GMEM_MOVEABLE, len);
                    memcpy(GlobalLock(hMem), url, len);
                    GlobalUnlock(hMem);
                    OpenClipboard(0);
                    EmptyClipboard();
                    SetClipboardData(CF_TEXT, hMem);
                    CloseClipboard();
                }
                break;
            }
            break;
        case WM_DESTROY:
            Exit();
            break;
        case WM_MYMESSAGE:
            switch(lParam)
            {
                case WM_RBUTTONUP:
                    ShowMenu(hWnd);
                    break;
                case WM_LBUTTONUP:
                    OpenInBrowser();
                    break;
                default:
                    return DefWindowProc(hWnd, message, wParam, lParam);
            };
            break;
        default:
            return DefWindowProc(hWnd, message, wParam, lParam);
    }
    return 0;
}

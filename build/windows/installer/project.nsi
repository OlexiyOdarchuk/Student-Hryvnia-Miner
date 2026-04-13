!pragma textin utf-8
Unicode true

####
## Wails defaults and macro includes
####
!include "wails_tools.nsh"
!include "LogicLib.nsh" 

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"
VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Інсталятор"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true

# --- НАЛАШТУВАННЯ ІНТЕРФЕЙСУ (Modern UI) ---
!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_ABORTWARNING # Попередження при скасуванні

# --- СТОРІНКИ ІНСТАЛЯТОРА ---
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY 
!insertmacro MUI_PAGE_INSTFILES

# Перекладений текст для галочки після встановлення
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Запустити ${INFO_PRODUCTNAME}"
!insertmacro MUI_PAGE_FINISH

# --- СТОРІНКИ ДЕІНСТАЛЯТОРА ---
!insertmacro MUI_UNPAGE_CONFIRM 
!insertmacro MUI_UNPAGE_INSTFILES

# УВІМКНЕННЯ УКРАЇНСЬКОЇ МОВИ
!insertmacro MUI_LANGUAGE "Ukrainian" 

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
ShowInstDetails show

# ==========================================
# ФУНКЦІЇ ІНСТАЛЯТОРА
# ==========================================
Function .onInit
    !insertmacro wails.checkArchitecture
    
    # Тихо вбиваємо процес перед оновленням
    nsExec::Exec 'taskkill /F /IM "${PRODUCT_EXECUTABLE}" /T'
FunctionEnd

Section "install"
    !insertmacro wails.setShellContext
    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    !insertmacro wails.writeUninstaller
SectionEnd

# ==========================================
# ФУНКЦІЇ ДЕІНСТАЛЯТОРА
# ==========================================
Function un.onInit
    # Тихо вбиваємо процес перед видаленням
    nsExec::Exec 'taskkill /F /IM "${PRODUCT_EXECUTABLE}" /T'

    # Перекладений запит на видалення конфігів
    MessageBox MB_YESNO|MB_ICONQUESTION "Чи бажаєте ви видалити ваші налаштування та дані гаманців?$\r$\n$\r$\n(Оберіть 'Ні', якщо плануєте перевстановити ${INFO_PRODUCTNAME})" IDNO keep_settings
    
    StrCpy $0 "DELETE"
    Goto done
    
keep_settings:
    StrCpy $0 "KEEP"

done:
FunctionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Видаляємо кеш WebView2

    # Видаляємо конфіги, якщо користувач погодився
    ${If} $0 == "DELETE"
        SetShellVarContext current
        RMDir /r "$AppData\SHMiner" 
        SetShellVarContext all
    ${EndIf}

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols
    !insertmacro wails.deleteUninstaller
SectionEnd
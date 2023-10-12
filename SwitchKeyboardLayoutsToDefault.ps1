Add-Type -AssemblyName 'System.Windows.Forms'
$WinUserLanguageList=Get-WinUserLanguageList
if ([System.Windows.Forms.InputLanguage]::CurrentInputLanguage.Culture.Name -ne $WinUserLanguageList[0].LanguageTag) {
 Set-WinUserLanguageList $WinUserLanguageList -Force
}
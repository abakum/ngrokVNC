cd /d %~dp0
attrib VERSION|find "A            "&&call :ver
attrib -a VERSION
goto :EOF

:ver
set /p VERSION=<VERSION
git tag v%VERSION%-lw
git push origin --tags
go-winres make --product-version=git-tag --file-version=git-tag
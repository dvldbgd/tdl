' [TODO] Add validation for user input
Dim username As String
username = Console.ReadLine()

' [NOTE] Just printing for now
Console.WriteLine("Hello, " & username)

' [FIXME] Crashes if input is null
' [HACK] For testing only – remove in production
If username = "admin" Then
    Console.WriteLine("Access granted.")
End If

' [OPTIMIZE] Reduce string concatenation calls
' [BUG] Case sensitivity issue in username check

' [DEPRECATED] Replace with GetUserInput() method


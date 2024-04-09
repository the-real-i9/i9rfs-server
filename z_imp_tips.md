## Implementation tips
- On the CLI app, when the user executes `i9rfs`,
  - if there is an auth token, use it to retrieve the user data for use in the session, depending on the validity of the token
  - if the token has expired then the client must type `i9rfs login uname={} pwd={}`, in which you use to execute a login procedure, and the client gets a new token
  - implement `i9rfs signup email={}` as usual.
  - The command line working directory where command is executed is of the pattern `i9rfs@{username}:~[/working/path/in/the/root]$`
- On the server-side, the `WorkPath` sent from the client starts from their username e.g. `/i9[/{...}]`.
- On the server-side, create a user account folder (i.e. user's root folder) in their username, on successful signup. The parent folder be ignored if it already exists.
  ```go
  cmd := exec.Command("mkdir", "-p", "i9FSHome/{username}")
  cmd.Run()
  ```
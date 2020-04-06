package sessions

import (
	"github.com/gorilla/sessions"
)

// Sessions stored in cookies. the byte array is a key to sign cookies, the package only accepts cookies signed with our key
var Store = sessions.NewCookieStore([]byte("secret-key"))
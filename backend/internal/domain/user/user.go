package user

type User struct {
	id       string
	fullname string
	email    string
	photoURL string // URL format
}

func NewUser(id, fullname, email, photoURL string) *User {
	return &User{
		id:       id,
		fullname: fullname,
		email:    email,
		photoURL: photoURL,
	}
}

func (u *User) GetID() string {
	return u.id
}

func (u *User) GetFullname() string {
	return u.fullname
}

func (u *User) GetEmail() string {
	return u.email
}

func (u *User) GetPhotoURL() string {
	return u.photoURL
}

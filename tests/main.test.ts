const HOST_URL = "http://localhost:8000"

const signupPath = HOST_URL + "/api/auth/signup"
const signinPath = HOST_URL + "/api/auth/signin"

const signoutPath = HOST_URL + "/api/app/signout"

const rfsPath = "ws://localhost:8000/rfs"

type UserT = {
  email: string
  username: string
  password: string
  sessionCookie: string
}

/* func containsDirs(dir ...string) td.TestDeep {
	containsList := make([]any, len(dir))

	for i, dirName := range dir {
		containsList[i] = td.Contains(td.SuperMapOf(map[string]any{
			"id":       td.Ignore(),
			"obj_type": "directory",
			"name":     dirName,
		}, nil))
	}

	return td.All(containsList...)
} */

/* func notContainsDirs(dir ...string) td.TestDeep {
	notContainsList := make([]any, len(dir))

	for i, dirName := range dir {
		notContainsList[i] = td.Not(td.Contains(td.SuperMapOf(map[string]any{
			"id":       td.Ignore(),
			"obj_type": "directory",
			"name":     dirName,
		}, nil)))
	}

	return td.All(notContainsList...)
} */

func TestUserAuthStory(t *testing.T) {
	t.Parallel()

	user := UserT{
		Email:    "suberu@gmail.com",
		Username: "suberu",
		Password: "sketeppy",
	}

	{
		t.Log("Action: user requests a new account")

		reqBody, err := makeReqBody(map[string]any{"email": user.Email})
		require.NoError(t, err)

		res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user.Email),
			}, nil))

		user.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user sends an incorrect email verf code")

		reqBody, err := makeReqBody(map[string]any{"code": "011111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, "Incorrect verification code! Check or Re-submit your email.", rb)
	}

	{
		t.Log("Action: user sends the correct email verification code")

		verfCode := os.Getenv("DUMMY_VERF_TOKEN")

		reqBody, err := makeReqBody(map[string]any{"code": verfCode})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("Your email, %s, has been verified!", user.Email),
			}, nil))

		user.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user submits her information")

		reqBody, err := makeReqBody(map[string]any{
			"username": user.Username,
			"password": user.Password,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"user": td.Ignore(),
				"msg":  "Signup success!",
			}, nil))

		user.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user signs out")

		req, err := http.NewRequest("GET", signoutPath, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}
	}

	{
		t.Log("Action: user signs in with incorrect credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user.Email,
			"password":        "millinix",
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusNotFound, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		require.Equal(t, "Incorrect email or password", rb)
	}

	{
		t.Log("Action: user signs in with correct credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user.Email,
			"password":        user.Password,
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": "Signin success!",
			}, nil))

		user.SessionCookie = res.Header.Get("Set-Cookie")
	}
}

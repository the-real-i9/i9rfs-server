import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as authValidators from "../validators/authValidators.ts"
import * as signupControllers from "../controllers/auth/signupControllers.ts"
import * as signinController from "../controllers/auth/signinController.ts"

const router = express.Router()

router.post(
  "/signup/request_new_account",
  ...authValidators.RequestNewAccount(),
  signupControllers.RequestNewAccountDoc,
  signupControllers.RequestNewAccount
)
router.post(
  "/signup/verify_email",
  ...authValidators.VerifyEmail(),
  authMiddlewares.SignupSession,
  signupControllers.VerifyEmailDoc,
  signupControllers.VerifyEmail
)
router.post(
  "/signup/register_user",
  ...authValidators.RegisterUser(),
  authMiddlewares.SignupSession,
  signupControllers.RegisterUserDoc,
  signupControllers.RegisterUser
)

router.post(
  "/signin",
  ...authValidators.Signin(),
  signinController.SigninDoc,
  signinController.Signin
)

export default router

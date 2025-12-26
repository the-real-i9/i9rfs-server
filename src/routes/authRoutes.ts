import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as authValidators from "../validators/authValidators.ts"
import * as signupControllers from "../controllers/auth/signupControllers.ts"
import * as signinController from "../controllers/auth/signinController.ts"

const router = express.Router()

router.post(
  "/signup/request_new_account",
  ...authValidators.RequestNewAccount(),
  signupControllers.RequestNewAccount
)
router.post(
  "/signup/verify_email",
  ...authValidators.VerifyEmail(),
  authMiddlewares.SignupSession,
  signupControllers.VerifyEmail
)
router.post(
  "/signup/register_user",
  ...authValidators.RegisterUser(),
  authMiddlewares.SignupSession,
  signupControllers.RegisterUser
)

router.post("/signin", ...authValidators.Signin(), signinController.Signin)

export default router

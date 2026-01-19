import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as appControllers from "../controllers/app/appControllers.ts"
import * as appValidators from "../validators/appValidators.ts"

const router = express.Router()

router.use(authMiddlewares.UserAuth)

router.post(
  "/uploads/authorize",
  ...appValidators.AuthorizeUpload(),
  appControllers.AuthorizeUploadDoc,
  appControllers.AuthorizeUpload
)

router.get("/signout", appControllers.SignoutDoc, appControllers.Signout)

export default router

import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as appControllers from "../controllers/app/appControllers.ts"
import * as appValidators from "../validators/appValidators.ts"

const router = express.Router()

router.use(authMiddlewares.UserAuth)

router.post(
  "/uploads/authorize",
  ...appValidators.AuthorizeUpload(),
  appControllers.AuthorizeUpload
)

router.post(
  "/uploads/create_file_object",
  ...appValidators.CreateFileObject(),
  appControllers.CreateFileObject
)

router.get("/signout", appControllers.Signout)

export default router

import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as appControllers from "../controllers/app/appControllers.ts"
import * as appValidators from "../validators/appValidators.ts"
import * as appMiddlewares from "../middlewares/appMiddlewares.ts"

const router = express.Router()

router.use(authMiddlewares.UserAuth)

router.get("/signout", appControllers.Signout)

router.post(
  "/uploads/authorize",
  ...appValidators.AuthorizeUpload(),
  appControllers.AuthorizeUpload
)
router.post(
  "/uploads/cloud_upload_complete",
  ...appValidators.CloudUploadComplete(),
  appMiddlewares.UploadSessionCUC,
  appControllers.CloudUploadComplete
)
router.post(
  "/uploads/create_file_object",
  ...appValidators.CreateFileObject(),
  appMiddlewares.UploadSessionCFO,
  appControllers.CreateFileObject
)

export default router

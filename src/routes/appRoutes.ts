import express from "express"
import * as authMiddlewares from "../middlewares/authMiddlewares.ts"
import * as appControllers from "../controllers/app/appControllers.ts"

const router = express.Router()

router.use(authMiddlewares.UserAuth)

router.get("/signout", appControllers.Signout)

export default router

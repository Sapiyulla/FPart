import { createBrowserRouter } from "react-router-dom";
import { AuthPage } from "../pages/Authorization";

export const router = createBrowserRouter([
    {
        path: "/",
        element: <AuthPage />
    }
])
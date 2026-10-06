import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Navigate, RouterProvider, createBrowserRouter } from "react-router-dom";

import App from "@/App";
import { ToastProvider } from "@/components/ToastProvider";
import "@/index.css";
import { ItemsPage } from "@/pages/ItemsPage";

const queryClient = new QueryClient();

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      {
        index: true,
        element: <Navigate to="/items" replace />,
      },
      {
        path: "items",
        element: <ItemsPage />,
      },
      {
        path: "items/new",
        element: <ItemsPage />,
      },
      {
        path: "items/:id/edit",
        element: <ItemsPage />,
      },
    ],
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  </StrictMode>,
);

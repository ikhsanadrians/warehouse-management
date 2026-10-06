import { Outlet } from "react-router-dom";

export default function App() {
  return (
    <main className="min-h-screen bg-muted/30 text-foreground">
      <div className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <Outlet />
      </div>
    </main>
  );
}

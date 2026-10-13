import { createRoot } from "react-dom/client";
import App from "./App";
import { setDesktopAPIBaseURL } from "./lib/api";
import "./index.css";

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetAPIBaseURL: () => Promise<string>;
        };
      };
    };
  }
}

async function start() {
  const root = document.getElementById("root");
  if (!root) throw new Error("The app root element is missing.");

  // In the desktop build the Go host supplies a private, random loopback port.
  // Vite preview remains available for frontend-only development.
  const desktop = window.go?.main?.App;
  if (desktop) {
    try {
      setDesktopAPIBaseURL(await desktop.GetAPIBaseURL());
    } catch (error) {
      root.innerHTML = "";
      const panel = document.createElement("main");
      panel.style.cssText = "max-width:640px;margin:12vh auto;padding:32px;font:16px/1.6 system-ui;color:#17243a";
      const heading = document.createElement("h1");
      heading.textContent = "Your local workspace could not start";
      const details = document.createElement("p");
      details.textContent = error instanceof Error ? error.message : "Restart the app and try again.";
      panel.append(heading, details);
      root.append(panel);
      return;
    }
  }

  createRoot(root).render(<App />);
}

void start();

import { useEffect, useLayoutEffect, useRef } from "react";

interface SSECallbacks {
  onUpdate: () => void;
  onFileChanged?: (fileId: string) => void;
}

declare global {
  interface Window {
    runtime?: {
      EventsOn: (eventName: string, callback: (...args: any[]) => void) => () => void;
      EventsOff: (eventName: string, ...callbacks: any[]) => void;
    };
  }
}

export function useSSE(callbacks: SSECallbacks) {
  const callbacksRef = useRef(callbacks);
  useLayoutEffect(() => {
    callbacksRef.current = callbacks;
  });

  useEffect(() => {
    let disposed = false;
    const unsubs: (() => void)[] = [];

    // Listen to Wails native events if running in desktop app
    if (typeof window !== "undefined" && window.runtime?.EventsOn) {
      const unsubUpdate = window.runtime.EventsOn("update", () => {
        if (!disposed) {
          callbacksRef.current.onUpdate();
        }
      });
      const unsubFileChanged = window.runtime.EventsOn("file-changed", (data: any) => {
        if (disposed) return;
        try {
          const parsed = typeof data === "string" ? JSON.parse(data) : data;
          if (parsed && typeof parsed.id === "string") {
            callbacksRef.current.onFileChanged?.(parsed.id);
          } else if (typeof data === "string") {
            callbacksRef.current.onFileChanged?.(data);
          }
        } catch {
          if (typeof data === "string") {
            callbacksRef.current.onFileChanged?.(data);
          }
        }
      });

      if (typeof unsubUpdate === "function") unsubs.push(unsubUpdate);
      if (typeof unsubFileChanged === "function") unsubs.push(unsubFileChanged);
    }

    let es: EventSource | null = null;
    let retryDelay = 1000;
    const maxRetryDelay = 30000;
    let serverPid: number | null = null;

    function connect() {
      if (disposed) return;

      try {
        es = new EventSource("/_/events");

        es.addEventListener("started", (e) => {
          try {
            const data = JSON.parse(e.data);
            if (typeof data.pid !== "number") return;
            if (serverPid !== null && data.pid !== serverPid) {
              window.location.reload();
              return;
            }
            serverPid = data.pid;
          } catch {
            // ignore
          }
        });

        es.addEventListener("update", () => {
          callbacksRef.current.onUpdate();
        });

        es.addEventListener("file-changed", (e) => {
          try {
            const data = JSON.parse(e.data);
            callbacksRef.current.onFileChanged?.(data.id);
          } catch {
            // ignore malformed data
          }
        });

        es.onopen = () => {
          retryDelay = 1000;
        };

        es.onerror = () => {
          es?.close();
          if (!disposed) {
            setTimeout(connect, retryDelay);
            retryDelay = Math.min(retryDelay * 2, maxRetryDelay);
          }
        };
      } catch {
        // ignore in environments without standard EventSource
      }
    }

    connect();

    return () => {
      disposed = true;
      es?.close();
      for (const unsub of unsubs) {
        unsub();
      }
    };
  }, []);
}

import { ImageResponse } from "next/og";

export const runtime = "edge";
export const alt = "Ship - Infrastructure CLI for AI Agents";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OGImage() {
  return new ImageResponse(
    (
      <div
        style={{
          background: "#0d1117",
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ fontSize: 96, marginBottom: 16 }}>🚀</div>
        <div
          style={{
            fontSize: 72,
            fontWeight: 700,
            color: "#e6edf3",
            marginBottom: 16,
          }}
        >
          Ship
        </div>
        <div
          style={{
            fontSize: 28,
            color: "#8b949e",
            textAlign: "center",
            maxWidth: 700,
          }}
        >
          Deploy from terminal. No dashboards. No context-switching.
        </div>
        <div
          style={{
            fontSize: 20,
            color: "#58a6ff",
            marginTop: 24,
          }}
        >
          Infrastructure CLI for AI Agents
        </div>
      </div>
    ),
    { ...size }
  );
}

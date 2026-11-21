import type React from "react";
import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {EuropeanSleeperLogo} from "../components/Logo";

export const DeckShell: React.FC<{readonly eyebrow: string; readonly children: React.ReactNode; readonly dark?: boolean}> = ({eyebrow, children, dark = true}) => {
  const frame = useCurrentFrame();
  const {durationInFrames} = useVideoConfig();
  return <AbsoluteFill style={{fontFamily: "Helvetica Neue, Helvetica, Arial, sans-serif", color: dark ? "#F8FAFC" : "#101720", backgroundColor: dark ? "#0B1120" : "#F7F3EB", backgroundImage: dark ? "radial-gradient(circle at 88% 8%, rgba(23,37,84,0.9), transparent 42%), linear-gradient(rgba(148,163,184,0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(148,163,184,0.05) 1px, transparent 1px)" : "radial-gradient(circle at 86% 10%, rgba(255,55,15,0.10), transparent 40%)", backgroundSize: "auto, 64px 64px, 64px 64px", opacity: interpolate(frame, [0, 15, durationInFrames - 16, durationInFrames - 1], [0.72, 1, 1, 0.85], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.bezier(0.16, 1, 0.3, 1)})}}><div style={{position: "absolute", left: 82, right: 82, top: 48, display: "flex", alignItems: "center", justifyContent: "space-between"}}><div style={{display: "flex", alignItems: "center", gap: 28}}><EuropeanSleeperLogo width={230} /><div style={{width: 1, height: 40, backgroundColor: dark ? "rgba(255,255,255,0.22)" : "rgba(16,23,32,0.18)"}} /><div style={{fontSize: 22, letterSpacing: 3.2, textTransform: "uppercase", fontWeight: 800, color: "#F59E0B"}}>{eyebrow}</div></div><div style={{fontSize: 20, letterSpacing: 2, color: dark ? "#94A3B8" : "#59616D"}}>ES INCIDENT COMMAND</div></div>{children}</AbsoluteFill>;
};

export const Rise: React.FC<{readonly delay?: number; readonly children: React.ReactNode; readonly style?: React.CSSProperties}> = ({delay = 0, children, style}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <div style={{opacity: interpolate(frame, [delay * fps, (delay + 0.65) * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"}), translate: interpolate(frame, [delay * fps, (delay + 0.65) * fps], ["0px 34px", "0px 0px"], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.bezier(0.16, 1, 0.3, 1)}), ...style}}>{children}</div>;
};

export const DeckCard: React.FC<{readonly children: React.ReactNode; readonly accent?: string}> = ({children, accent = "#334155"}) => <div style={{backgroundColor: "rgba(19,28,49,0.92)", border: `1px solid ${accent}`, borderRadius: 24, padding: "30px 32px", boxShadow: "0 20px 50px rgba(0,0,0,0.28)"}}>{children}</div>;

export const Pill: React.FC<{readonly children: React.ReactNode; readonly color: string}> = ({children, color}) => <span style={{display: "inline-block", padding: "9px 15px", borderRadius: 999, fontSize: 19, fontWeight: 800, letterSpacing: 1.2, textTransform: "uppercase", color, border: `1px solid ${color}`, backgroundColor: `${color}20`}}>{children}</span>;

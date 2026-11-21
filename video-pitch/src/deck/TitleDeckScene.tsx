import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {DeckShell, Pill, Rise} from "./DeckShell";

export const TitleDeckScene: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <AbsoluteFill><DeckShell eyebrow="Operations AI Proof of Concept"><div style={{position: "absolute", left: 95, right: 95, top: 245}}><Rise><div style={{fontSize: 122, lineHeight: 0.98, fontWeight: 800, letterSpacing: -6}}>ES Incident<br /><span style={{color: "#38BDF8"}}>Command</span></div></Rise><Rise delay={0.45}><div style={{fontSize: 39, lineHeight: 1.4, color: "#CBD5E1", marginTop: 34, maxWidth: 1200}}>Neuro-symbolic decision support for last-minute night-train disruptions.</div></Rise><Rise delay={0.9} style={{marginTop: 54, display: "flex", gap: 16}}><Pill color="#10B981">Go rail guardrails</Pill><Pill color="#F59E0B">Laya System 1</Pill><Pill color="#38BDF8">Gemini System 2</Pill><Pill color="#A7F3D0">Wails + React</Pill></Rise></div><div style={{position: "absolute", right: 95, bottom: 80, width: 280, height: 10, borderRadius: 99, backgroundColor: "#F59E0B", scale: interpolate(frame, [1.1 * fps, 2.1 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.spring({damping: 180}), output: "perceptual-scale"}), transformOrigin: "right center"}} /></DeckShell></AbsoluteFill>;
};

import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {Backdrop} from "../components/Backdrop";
import {SceneChrome} from "../components/SceneChrome";

const Metric = ({value, label, delay}: {value: string; label: string; delay: number}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <div style={{width: 430, padding: "38px 42px", borderRadius: 26, backgroundColor: "rgba(255,255,255,0.82)", border: "1px solid rgba(16,23,32,0.12)", boxShadow: "0 18px 48px rgba(16,23,32,0.08)", opacity: interpolate(frame, [delay * fps, (delay + 0.7) * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"}), translate: interpolate(frame, [delay * fps, (delay + 0.7) * fps], ["0px 38px", "0px 0px"], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.bezier(0.16, 1, 0.3, 1)})}}><div style={{fontSize: 94, fontWeight: 800, color: "#FF370F", letterSpacing: -4}}>{value}</div><div style={{fontSize: 31, lineHeight: 1.25, color: "#4D5662", marginTop: 10}}>{label}</div></div>;
};

export const ProblemScene: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <AbsoluteFill><Backdrop /><SceneChrome section="The operational reality"><div style={{position: "absolute", left: 90, right: 90, top: 205}}><div style={{fontSize: 78, lineHeight: 1.05, fontWeight: 760, letterSpacing: -3, maxWidth: 1200, opacity: interpolate(frame, [0, 0.8 * fps], [0, 1], {extrapolateRight: "clamp"})}}>A live railway plan trapped in a static spreadsheet.</div><div style={{display: "flex", gap: 28, marginTop: 72}}><Metric value="~70" label="leased and owned carriages" delay={0.5} /><Metric value="3" label="night-train lines from Brussels" delay={0.9} /><Metric value="1" label="colour-coded operational plan" delay={1.3} /></div><div style={{marginTop: 52, display: "flex", gap: 18, alignItems: "center", fontSize: 34, color: "#4D5662", opacity: interpolate(frame, [1.6 * fps, 2.5 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"})}}><span style={{padding: "12px 22px", borderRadius: 99, backgroundColor: "#101720", color: "white", fontWeight: 700}}>No change log</span><span>·</span><span style={{padding: "12px 22px", borderRadius: 99, backgroundColor: "#101720", color: "white", fontWeight: 700}}>No automatic rule check</span><span>·</span><span style={{padding: "12px 22px", borderRadius: 99, backgroundColor: "#101720", color: "white", fontWeight: 700}}>Multilingual free text</span></div></div></SceneChrome></AbsoluteFill>;
};

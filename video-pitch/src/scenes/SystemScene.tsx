import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {Backdrop} from "../components/Backdrop";
import {SceneChrome} from "../components/SceneChrome";

const Stage = ({number, title, copy, delay, color}: {number: string; title: string; copy: string; delay: number; color: string}) => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <div style={{width: 480, minHeight: 390, padding: "38px 38px 34px", borderRadius: 30, backgroundColor: "rgba(255,255,255,0.93)", border: "1px solid rgba(16,23,32,0.12)", boxShadow: "0 20px 60px rgba(0,0,0,0.18)", opacity: interpolate(frame, [delay * fps, (delay + 0.8) * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"}), scale: interpolate(frame, [delay * fps, (delay + 0.8) * fps], [0.92, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.spring({damping: 180}), output: "perceptual-scale"})}}><div style={{width: 68, height: 68, borderRadius: 99, backgroundColor: color, color: "white", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 30, fontWeight: 800}}>{number}</div><div style={{fontSize: 43, fontWeight: 780, color: "#101720", marginTop: 28, letterSpacing: -1.5}}>{title}</div><div style={{fontSize: 27, lineHeight: 1.42, color: "#59616D", marginTop: 22}}>{copy}</div></div>;
};

export const SystemScene: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return <AbsoluteFill><Backdrop dark /><SceneChrome section="The decision pipeline" dark><div style={{position: "absolute", left: 90, right: 90, top: 190}}><div style={{fontSize: 76, fontWeight: 760, letterSpacing: -3}}>Evidence first. Rules second. AI last.</div><div style={{display: "flex", gap: 34, marginTop: 72, alignItems: "center"}}><Stage number="1" title="Evidence" copy="Preserve raw cells, colours, comments, timestamps and provenance. Unknown stays unknown." delay={0.4} color="#466D7A" /><div style={{fontSize: 70, color: "#FF7555", opacity: interpolate(frame, [1 * fps, 1.5 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"})}}>→</div><Stage number="2" title="Rules" copy="Block incompatible countries, train length, doors, high voltage, shunting and maintenance conflicts." delay={1.0} color="#D92700" /><div style={{fontSize: 70, color: "#FF7555", opacity: interpolate(frame, [1.6 * fps, 2.1 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"})}}>→</div><Stage number="3" title="Rank" copy="Jev scores only prevalidated actions with typed answers. It never invents the options." delay={1.6} color="#8B4B9C" /></div><div style={{marginTop: 44, fontSize: 28, color: "#D9E0E6", opacity: interpolate(frame, [2.2 * fps, 3.1 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"})}}>If a blocking fact is missing: <span style={{color: "#FF8A6E", fontWeight: 800}}>UNVERIFIED → escalate</span></div></div></SceneChrome></AbsoluteFill>;
};

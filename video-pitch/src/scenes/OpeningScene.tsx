import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {Backdrop} from "../components/Backdrop";
import {EuropeanSleeperLogo} from "../components/Logo";

export const OpeningScene: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  return (
    <AbsoluteFill style={{fontFamily: "Helvetica Neue, Helvetica, Arial, sans-serif", color: "#101720"}}>
      <Backdrop />
      <div style={{position: "absolute", left: 110, right: 110, top: 90, display: "flex", justifyContent: "space-between", alignItems: "center"}}>
        <div style={{opacity: interpolate(frame, [0, 0.8 * fps], [0, 1], {extrapolateRight: "clamp"}), translate: interpolate(frame, [0, 0.8 * fps], ["-40px 0px", "0px 0px"], {extrapolateRight: "clamp", easing: Easing.bezier(0.16, 1, 0.3, 1)})}}>
          <EuropeanSleeperLogo width={500} />
        </div>
        <div style={{fontSize: 24, letterSpacing: 3.5, textTransform: "uppercase", fontWeight: 700, color: "#C52C0A"}}>Hack the Rails · Final</div>
      </div>
      <div style={{position: "absolute", left: 110, bottom: 112, width: 1500}}>
        <div style={{fontSize: 96, lineHeight: 1.02, fontWeight: 750, letterSpacing: -4, opacity: interpolate(frame, [0.7 * fps, 1.7 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"}), translate: interpolate(frame, [0.7 * fps, 1.7 * fps], ["0px 55px", "0px 0px"], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.bezier(0.16, 1, 0.3, 1)})}}>
          When one carriage fails,
          <br />the whole night plan moves.
        </div>
        <div style={{marginTop: 36, fontSize: 36, lineHeight: 1.35, maxWidth: 1100, color: "#4D5662", opacity: interpolate(frame, [1.6 * fps, 2.4 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp"})}}>
          An explainable decision assistant for rolling-stock disruption triage.
        </div>
      </div>
      <div style={{position: "absolute", right: 110, bottom: 118, width: 210, height: 14, borderRadius: 99, backgroundColor: "#FF370F", scale: interpolate(frame, [1.2 * fps, 2.2 * fps], [0, 1], {extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.spring({damping: 180}), output: "perceptual-scale"}), transformOrigin: "right center"}} />
    </AbsoluteFill>
  );
};

import type React from "react";
import {AbsoluteFill, Easing, interpolate, useCurrentFrame, useVideoConfig} from "remotion";

export const Backdrop: React.FC<{readonly dark?: boolean}> = ({dark = false}) => {
  const frame = useCurrentFrame();
  const {durationInFrames} = useVideoConfig();
  return (
    <AbsoluteFill
      style={{
        backgroundColor: dark ? "#101720" : "#F7F3EB",
        backgroundImage: dark
          ? "radial-gradient(circle at 78% 16%, rgba(255,55,15,0.16), transparent 32%), linear-gradient(rgba(255,255,255,0.035) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.035) 1px, transparent 1px)"
          : "radial-gradient(circle at 82% 18%, rgba(255,55,15,0.10), transparent 31%), linear-gradient(rgba(16,23,32,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(16,23,32,0.04) 1px, transparent 1px)",
        backgroundSize: "auto, 54px 54px, 54px 54px",
        opacity: interpolate(frame, [0, 18, durationInFrames - 20, durationInFrames - 1], [0.7, 1, 1, 0.88], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(0.16, 1, 0.3, 1),
        }),
      }}
    />
  );
};

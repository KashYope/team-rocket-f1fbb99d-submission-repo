import type React from "react";
import {AbsoluteFill} from "remotion";
import {EuropeanSleeperLogo} from "./Logo";

export const SceneChrome: React.FC<{readonly section: string; readonly dark?: boolean; readonly children: React.ReactNode}> = ({section, dark = false, children}) => (
  <AbsoluteFill
    style={{fontFamily: "Helvetica Neue, Helvetica, Arial, sans-serif", color: dark ? "#FFFFFF" : "#101720"}}
  >
    <div style={{position: "absolute", left: 82, top: 54, display: "flex", alignItems: "center", gap: 28}}>
      <EuropeanSleeperLogo width={235} />
      <div style={{width: 1, height: 42, backgroundColor: dark ? "rgba(255,255,255,0.24)" : "rgba(16,23,32,0.18)"}} />
      <div style={{fontSize: 24, letterSpacing: 2.8, textTransform: "uppercase", fontWeight: 700, color: dark ? "#F6B59F" : "#C52C0A"}}>{section}</div>
    </div>
    {children}
  </AbsoluteFill>
);

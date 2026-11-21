import type React from "react";
import {Img, staticFile} from "remotion";

export const EuropeanSleeperLogo: React.FC<{readonly width?: number}> = ({width = 430}) => (
  <Img
    name="European Sleeper logo"
    src={staticFile("assets/european-sleeper-logo.png")}
    style={{width, height: "auto", objectFit: "contain"}}
  />
);

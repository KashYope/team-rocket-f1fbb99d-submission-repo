import {Audio} from "@remotion/media";
import {Series, staticFile, useVideoConfig} from "remotion";
import {OpeningScene} from "./scenes/OpeningScene";
import {ProblemScene} from "./scenes/ProblemScene";
import {CascadeScene} from "./scenes/CascadeScene";
import {SystemScene} from "./scenes/SystemScene";
import {DashboardScene} from "./scenes/DashboardScene";
import {GuardrailsScene} from "./scenes/GuardrailsScene";
import {ImpactScene} from "./scenes/ImpactScene";

export const PitchVideo: React.FC = () => {
  const {fps} = useVideoConfig();
  return <><Series><Series.Sequence name="Opening" durationInFrames={10 * fps} premountFor={fps}><OpeningScene /></Series.Sequence><Series.Sequence name="Problem" durationInFrames={16 * fps} premountFor={fps}><ProblemScene /></Series.Sequence><Series.Sequence name="Cascade" durationInFrames={17 * fps} premountFor={fps}><CascadeScene /></Series.Sequence><Series.Sequence name="System" durationInFrames={23 * fps} premountFor={fps}><SystemScene /></Series.Sequence><Series.Sequence name="Dashboard" durationInFrames={24 * fps} premountFor={fps}><DashboardScene /></Series.Sequence><Series.Sequence name="Guardrails" durationInFrames={17 * fps} premountFor={fps}><GuardrailsScene /></Series.Sequence><Series.Sequence name="Impact" durationInFrames={13 * fps} premountFor={fps}><ImpactScene /></Series.Sequence></Series><Audio
    name="English narration"
    src={staticFile("audio/pitch-narration.mp3")}
    premountFor={fps}
    volume={1}
    hidden
  /></>;
};

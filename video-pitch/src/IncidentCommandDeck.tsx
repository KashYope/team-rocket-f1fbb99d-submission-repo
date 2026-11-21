import {Series, useVideoConfig} from "remotion";
import {TitleDeckScene} from "./deck/TitleDeckScene";
import {ProblemDeckScene} from "./deck/ProblemDeckScene";
import {SafetyDeckScene} from "./deck/SafetyDeckScene";
import {ArchitectureDeckScene} from "./deck/ArchitectureDeckScene";
import {SystemOneDeckScene} from "./deck/SystemOneDeckScene";
import {SystemTwoDeckScene} from "./deck/SystemTwoDeckScene";
import {ScenariosDeckScene} from "./deck/ScenariosDeckScene";
import {HumanControlDeckScene} from "./deck/HumanControlDeckScene";
import {ClosingDeckScene} from "./deck/ClosingDeckScene";

export const IncidentCommandDeck: React.FC = () => {
  const {fps} = useVideoConfig();
  return <Series><Series.Sequence name="Title" durationInFrames={12 * fps} premountFor={fps}><TitleDeckScene /></Series.Sequence><Series.Sequence name="Problem" durationInFrames={16 * fps} premountFor={fps}><ProblemDeckScene /></Series.Sequence><Series.Sequence name="Safety trap" durationInFrames={14 * fps} premountFor={fps}><SafetyDeckScene /></Series.Sequence><Series.Sequence name="Architecture" durationInFrames={18 * fps} premountFor={fps}><ArchitectureDeckScene /></Series.Sequence><Series.Sequence name="System 1" durationInFrames={16 * fps} premountFor={fps}><SystemOneDeckScene /></Series.Sequence><Series.Sequence name="System 2" durationInFrames={16 * fps} premountFor={fps}><SystemTwoDeckScene /></Series.Sequence><Series.Sequence name="Scenarios" durationInFrames={14 * fps} premountFor={fps}><ScenariosDeckScene /></Series.Sequence><Series.Sequence name="Human control" durationInFrames={8 * fps} premountFor={fps}><HumanControlDeckScene /></Series.Sequence><Series.Sequence name="Closing" durationInFrames={6 * fps} premountFor={fps}><ClosingDeckScene /></Series.Sequence></Series>;
};

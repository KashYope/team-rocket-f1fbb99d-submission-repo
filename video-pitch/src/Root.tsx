import {Composition, Folder} from "remotion";
import {PitchVideo} from "./PitchVideo";
import {OpeningScene} from "./scenes/OpeningScene";
import {ProblemScene} from "./scenes/ProblemScene";
import {CascadeScene} from "./scenes/CascadeScene";
import {SystemScene} from "./scenes/SystemScene";
import {DashboardScene} from "./scenes/DashboardScene";
import {GuardrailsScene} from "./scenes/GuardrailsScene";
import {ImpactScene} from "./scenes/ImpactScene";
import {IncidentCommandDeck} from "./IncidentCommandDeck";
import {TitleDeckScene} from "./deck/TitleDeckScene";
import {ProblemDeckScene} from "./deck/ProblemDeckScene";
import {SafetyDeckScene} from "./deck/SafetyDeckScene";
import {ArchitectureDeckScene} from "./deck/ArchitectureDeckScene";
import {SystemOneDeckScene} from "./deck/SystemOneDeckScene";
import {SystemTwoDeckScene} from "./deck/SystemTwoDeckScene";
import {ScenariosDeckScene} from "./deck/ScenariosDeckScene";
import {HumanControlDeckScene} from "./deck/HumanControlDeckScene";
import {ClosingDeckScene} from "./deck/ClosingDeckScene";

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Folder name="Scenes">
        <Composition id="Opening" component={OpeningScene} durationInFrames={300} fps={30} width={1920} height={1080} />
        <Composition id="Problem" component={ProblemScene} durationInFrames={480} fps={30} width={1920} height={1080} />
        <Composition id="Cascade" component={CascadeScene} durationInFrames={510} fps={30} width={1920} height={1080} />
        <Composition id="System" component={SystemScene} durationInFrames={690} fps={30} width={1920} height={1080} />
        <Composition id="Dashboard" component={DashboardScene} durationInFrames={720} fps={30} width={1920} height={1080} />
        <Composition id="Guardrails" component={GuardrailsScene} durationInFrames={510} fps={30} width={1920} height={1080} />
        <Composition id="Impact" component={ImpactScene} durationInFrames={390} fps={30} width={1920} height={1080} />
      </Folder>
      <Composition
        id="EuropeanSleeperPitch"
        component={PitchVideo}
        durationInFrames={3600}
        fps={30}
        width={1920}
        height={1080}
      />
      <Folder name="Incident-Command-Scenes">
        <Composition id="IC-Title" component={TitleDeckScene} durationInFrames={360} fps={30} width={1920} height={1080} />
        <Composition id="IC-Problem" component={ProblemDeckScene} durationInFrames={480} fps={30} width={1920} height={1080} />
        <Composition id="IC-Safety" component={SafetyDeckScene} durationInFrames={420} fps={30} width={1920} height={1080} />
        <Composition id="IC-Architecture" component={ArchitectureDeckScene} durationInFrames={540} fps={30} width={1920} height={1080} />
        <Composition id="IC-System-One" component={SystemOneDeckScene} durationInFrames={480} fps={30} width={1920} height={1080} />
        <Composition id="IC-System-Two" component={SystemTwoDeckScene} durationInFrames={480} fps={30} width={1920} height={1080} />
        <Composition id="IC-Scenarios" component={ScenariosDeckScene} durationInFrames={420} fps={30} width={1920} height={1080} />
        <Composition id="IC-Human-Control" component={HumanControlDeckScene} durationInFrames={240} fps={30} width={1920} height={1080} />
        <Composition id="IC-Closing" component={ClosingDeckScene} durationInFrames={180} fps={30} width={1920} height={1080} />
      </Folder>
      <Composition
        id="ESIncidentCommandDeck"
        component={IncidentCommandDeck}
        durationInFrames={3600}
        fps={30}
        width={1920}
        height={1080}
      />
    </>
  );
};

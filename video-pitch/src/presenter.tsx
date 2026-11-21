import React, {useCallback, useEffect, useRef, useState} from "react";
import {createRoot} from "react-dom/client";
import {Player, type PlayerRef} from "@remotion/player";
import {IncidentCommandDeck} from "./IncidentCommandDeck";
import "./presenter.css";

const FPS = 30;
const ENTRY_PLAY_TIME_MS = 2200;

const chapters = [
  {label: "Opening", frame: 0},
  {label: "Operational reality", frame: 360},
  {label: "AI safety trap", frame: 840},
  {label: "Architecture", frame: 1260},
  {label: "System 1", frame: 1800},
  {label: "System 2", frame: 2280},
  {label: "Stress tests", frame: 2760},
  {label: "Human control", frame: 3180},
  {label: "Live demo", frame: 3420},
] as const;

const Presenter: React.FC = () => {
  const playerRef = useRef<PlayerRef>(null);
  const pauseTimer = useRef<number | null>(null);
  const [chapterIndex, setChapterIndex] = useState(0);

  const showChapter = useCallback((nextIndex: number) => {
    const boundedIndex = Math.max(0, Math.min(chapters.length - 1, nextIndex));
    const player = playerRef.current;
    if (!player) return;

    if (pauseTimer.current !== null) window.clearTimeout(pauseTimer.current);
    setChapterIndex(boundedIndex);
    player.pause();
    player.seekTo(chapters[boundedIndex].frame);
    player.play();
    pauseTimer.current = window.setTimeout(() => {
      player.pause();
      pauseTimer.current = null;
    }, ENTRY_PLAY_TIME_MS);
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (["ArrowRight", "PageDown", " ", "Enter"].includes(event.key)) {
        event.preventDefault();
        showChapter(chapterIndex + 1);
      } else if (["ArrowLeft", "PageUp", "Backspace"].includes(event.key)) {
        event.preventDefault();
        showChapter(chapterIndex - 1);
      } else if (event.key === "Home") {
        event.preventDefault();
        showChapter(0);
      } else if (event.key === "End") {
        event.preventDefault();
        showChapter(chapters.length - 1);
      } else if (event.key.toLowerCase() === "f") {
        event.preventDefault();
        void playerRef.current?.requestFullscreen();
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      if (pauseTimer.current !== null) window.clearTimeout(pauseTimer.current);
    };
  }, [chapterIndex, showChapter]);

  useEffect(() => {
    const timer = window.setTimeout(() => showChapter(0), 100);
    return () => window.clearTimeout(timer);
  }, [showChapter]);

  return (
    <main className="presenter-shell">
      <div className="player-frame">
        <Player
          ref={playerRef}
          component={IncidentCommandDeck}
          durationInFrames={3600}
          compositionWidth={1920}
          compositionHeight={1080}
          fps={FPS}
          controls={false}
          autoPlay={false}
          clickToPlay={false}
          doubleClickToFullscreen
          spaceKeyToPlayOrPause={false}
          acknowledgeRemotionLicense
          style={{width: "100%", aspectRatio: "16 / 9"}}
        />
      </div>

      <nav className="presenter-controls" aria-label="Presentation controls">
        <button
          type="button"
          className="nav-button"
          disabled={chapterIndex === 0}
          onClick={() => showChapter(chapterIndex - 1)}
        >
          ← Previous
        </button>

        <div className="chapter-status" aria-live="polite">
          <strong>{chapters[chapterIndex].label}</strong>
          <span>{chapterIndex + 1} / {chapters.length}</span>
        </div>

        <div className="chapter-dots" aria-label="Choose chapter">
          {chapters.map((chapter, index) => (
            <button
              key={chapter.label}
              type="button"
              className={index === chapterIndex ? "chapter-dot active" : "chapter-dot"}
              aria-label={`Open ${chapter.label}`}
              aria-current={index === chapterIndex ? "step" : undefined}
              onClick={() => showChapter(index)}
            />
          ))}
        </div>

        <button
          type="button"
          className="nav-button primary"
          disabled={chapterIndex === chapters.length - 1}
          onClick={() => showChapter(chapterIndex + 1)}
        >
          Next →
        </button>
      </nav>

      <p className="keyboard-help">
        <kbd>←</kbd><kbd>→</kbd> chapters · <kbd>Space</kbd> next · <kbd>F</kbd> fullscreen
      </p>
    </main>
  );
};

const root = document.getElementById("root");
if (!root) throw new Error("Missing #root element");
createRoot(root).render(<Presenter />);

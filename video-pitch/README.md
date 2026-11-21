# European Sleeper incident-command presentations

Two English Remotion presentations for the Hack the Rails final.

## Preview

```bash
npm install
npm run dev -- --port=48080
```

Open `http://localhost:48080/ESIncidentCommandDeck` for the silent presentation.

For manual chapter control, run `npm run present`, then open
`http://localhost:48100/presenter.html`. Use Right Arrow, Space or Enter for the
next chapter; Left Arrow or Backspace for the previous chapter; F for fullscreen.

## Silent composition

- Resolution: 1920 × 1080
- Frame rate: 30 fps
- Duration: 3,600 frames, exactly 2:00
- Audio: none
- Data shown: synthetic only

The nine scenes are registered separately under `Incident-Command-Scenes` and assembled in `ESIncidentCommandDeck`. It incorporates the supplied ESICA architecture visualization.

The earlier narrated composition remains available as `EuropeanSleeperPitch`.

## Export

```bash
npx remotion render ESIncidentCommandDeck out/es-incident-command-deck.mp4
```

The supplied source assets are copied to `public/assets/european-sleeper-logo.png` and `public/assets/es-incident-command-architecture.jpeg`.

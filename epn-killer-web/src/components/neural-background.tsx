export const NeuralBackground = () => {
  return (
    <div className="neural-bg">
      {/* Static ambient glow — no canvas, no animation loop */}
      <div
        className="absolute inset-0"
        style={{
          background: `
            radial-gradient(ellipse 60% 50% at 20% 30%, rgba(59,130,246,0.08) 0%, transparent 70%),
            radial-gradient(ellipse 50% 40% at 75% 60%, rgba(139,92,246,0.07) 0%, transparent 70%),
            radial-gradient(ellipse 40% 50% at 50% 80%, rgba(99,102,241,0.05) 0%, transparent 70%)
          `,
        }}
      />
    </div>
  );
};

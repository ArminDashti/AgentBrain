/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        display: ['Fraunces', 'Georgia', 'Times New Roman', 'serif'],
        sans: ['Figtree', 'Segoe UI', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['IBM Plex Mono', 'ui-monospace', 'Cascadia Code', 'Consolas', 'monospace'],
      },
      colors: {
        border: 'var(--border)',
        background: 'var(--background)',
        foreground: 'var(--foreground)',
        card: 'var(--card)',
        primary: { DEFAULT: 'var(--primary)', foreground: 'var(--primary-foreground)' },
        muted: { DEFAULT: 'var(--muted)', foreground: 'var(--muted-foreground)' },
        danger: 'var(--danger)',
        success: 'var(--success)',
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}

import { defineConfig } from '@unocss/vite';
import { presetWind4 } from '@unocss/preset-wind4';

const c = (name: string) => `hsl(var(--${name}))`;

export default defineConfig({
  presets: [presetWind4({ dark: 'class' })],
  theme: {
    colors: {
      background: c('background'),
      foreground: c('foreground'),
      border: c('border'),
      input: c('input'),
      ring: c('ring'),
      card: { DEFAULT: c('card'), foreground: c('card-foreground') },
      popover: { DEFAULT: c('popover'), foreground: c('popover-foreground') },
      primary: { DEFAULT: c('primary'), foreground: c('primary-foreground') },
      secondary: { DEFAULT: c('secondary'), foreground: c('secondary-foreground') },
      muted: { DEFAULT: c('muted'), foreground: c('muted-foreground') },
      accent: { DEFAULT: c('accent'), foreground: c('accent-foreground') },
      destructive: { DEFAULT: c('destructive'), foreground: c('destructive-foreground') },
    },
    radius: {
      lg: 'var(--radius)',
      md: 'calc(var(--radius) - 2px)',
      sm: 'calc(var(--radius) - 4px)',
    },
  },
});

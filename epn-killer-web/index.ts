// Entry point for Expo web builds
import { registerRootComponent } from 'expo';
import App from './src/App';

// registerRootComponent calls AppRegistry.registerComponent('main', () => App);
// For web builds, Expo will use this entry point
registerRootComponent(App);

import { useEffect, useRef } from "react";

const vertexShaderSource = `
attribute vec2 aPosition;
varying vec2 vUv;

void main() {
  vUv = vec2(aPosition.x * 0.5 + 0.5, 0.5 - aPosition.y * 0.5);
  gl_Position = vec4(aPosition, 0.0, 1.0);
}
`;

const fragmentShaderSource = `
precision mediump float;

uniform sampler2D uTexture;
uniform vec2 uTexel;
uniform float uSharpness;
uniform float uSaturation;
uniform float uBrightness;
uniform float uContrast;

varying vec2 vUv;

vec3 adjustColor(vec3 color) {
  float luma = dot(color, vec3(0.2126, 0.7152, 0.0722));
  color = mix(vec3(luma), color, uSaturation);
  color *= uBrightness;
  color = (color - 0.5) * uContrast + 0.5;
  return color;
}

void main() {
  vec3 center = texture2D(uTexture, vUv).rgb;
  vec3 up = texture2D(uTexture, vUv + vec2(0.0, -uTexel.y)).rgb;
  vec3 down = texture2D(uTexture, vUv + vec2(0.0, uTexel.y)).rgb;
  vec3 left = texture2D(uTexture, vUv + vec2(-uTexel.x, 0.0)).rgb;
  vec3 right = texture2D(uTexture, vUv + vec2(uTexel.x, 0.0)).rgb;

  vec3 minColor = min(center, min(min(up, down), min(left, right)));
  vec3 maxColor = max(center, max(max(up, down), max(left, right)));
  vec3 range = min(minColor, 1.0 - maxColor);
  float adaptive = clamp(max(max(range.r, range.g), range.b) / max(max(maxColor.r, maxColor.g), max(maxColor.b, 0.001)), 0.0, 1.0);
  float weight = -sqrt(adaptive) * uSharpness * 0.125;
  vec3 sharpened = (center + weight * (up + down + left + right)) / (1.0 + 4.0 * weight);

  gl_FragColor = vec4(adjustColor(clamp(sharpened, 0.0, 1.0)), 1.0);
}
`;

interface VideoCasCanvasProps {
  readonly video: HTMLVideoElement | null;
  readonly width: number;
  readonly height: number;
  readonly sharpness: number;
  readonly saturation: number;
  readonly brightness: number;
  readonly contrast: number;
  readonly onReady?: () => void;
  readonly onUnavailable?: () => void;
}

function createShader(gl: WebGLRenderingContext, type: number, source: string) {
  const shader = gl.createShader(type);
  if (!shader) return null;

  gl.shaderSource(shader, source);
  gl.compileShader(shader);

  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    console.warn("CAS shader compile failed", gl.getShaderInfoLog(shader));
    gl.deleteShader(shader);
    return null;
  }

  return shader;
}

function createProgram(gl: WebGLRenderingContext) {
  const vertexShader = createShader(gl, gl.VERTEX_SHADER, vertexShaderSource);
  const fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, fragmentShaderSource);
  if (!vertexShader || !fragmentShader) return null;

  const program = gl.createProgram();
  if (!program) return null;

  gl.attachShader(program, vertexShader);
  gl.attachShader(program, fragmentShader);
  gl.linkProgram(program);

  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    console.warn("CAS program link failed", gl.getProgramInfoLog(program));
    gl.deleteProgram(program);
    return null;
  }

  return program;
}

export default function VideoCasCanvas({
  video,
  width,
  height,
  sharpness,
  saturation,
  brightness,
  contrast,
  onReady,
  onUnavailable,
}: VideoCasCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const settingsRef = useRef({ sharpness, saturation, brightness, contrast });

  useEffect(() => {
    settingsRef.current = { sharpness, saturation, brightness, contrast };
  }, [brightness, contrast, saturation, sharpness]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !video || width <= 0 || height <= 0) return;

    const gl = canvas.getContext("webgl", {
      alpha: false,
      antialias: false,
      depth: false,
      preserveDrawingBuffer: false,
      stencil: false,
    });
    if (!gl) {
      onUnavailable?.();
      return;
    }

    const program = createProgram(gl);
    if (!program) {
      onUnavailable?.();
      return;
    }

    const buffer = gl.createBuffer();
    const texture = gl.createTexture();
    if (!buffer || !texture) {
      onUnavailable?.();
      return;
    }

    gl.useProgram(program);
    onReady?.();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(
      gl.ARRAY_BUFFER,
      new Float32Array([-1, -1, 1, -1, -1, 1, -1, 1, 1, -1, 1, 1]),
      gl.STATIC_DRAW,
    );

    const position = gl.getAttribLocation(program, "aPosition");
    gl.enableVertexAttribArray(position);
    gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0);

    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

    const textureUniform = gl.getUniformLocation(program, "uTexture");
    const texelUniform = gl.getUniformLocation(program, "uTexel");
    const sharpnessUniform = gl.getUniformLocation(program, "uSharpness");
    const saturationUniform = gl.getUniformLocation(program, "uSaturation");
    const brightnessUniform = gl.getUniformLocation(program, "uBrightness");
    const contrastUniform = gl.getUniformLocation(program, "uContrast");

    gl.uniform1i(textureUniform, 0);

    let animationFrame = 0;
    const render = () => {
      animationFrame = requestAnimationFrame(render);

      if (video.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) return;

      const dpr = window.devicePixelRatio || 1;
      const canvasWidth = Math.max(1, Math.round(width * dpr));
      const canvasHeight = Math.max(1, Math.round(height * dpr));
      if (canvas.width !== canvasWidth || canvas.height !== canvasHeight) {
        canvas.width = canvasWidth;
        canvas.height = canvasHeight;
      }

      gl.viewport(0, 0, canvas.width, canvas.height);
      gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, video);

      const current = settingsRef.current;
      gl.uniform2f(
        texelUniform,
        1 / Math.max(1, video.videoWidth),
        1 / Math.max(1, video.videoHeight),
      );
      gl.uniform1f(sharpnessUniform, current.sharpness);
      gl.uniform1f(saturationUniform, current.saturation);
      gl.uniform1f(brightnessUniform, current.brightness);
      gl.uniform1f(contrastUniform, current.contrast);
      gl.drawArrays(gl.TRIANGLES, 0, 6);
    };

    render();

    return () => {
      cancelAnimationFrame(animationFrame);
      gl.deleteTexture(texture);
      gl.deleteBuffer(buffer);
      gl.deleteProgram(program);
    };
  }, [height, onReady, onUnavailable, video, width]);

  return <canvas ref={canvasRef} className="pointer-events-none absolute inset-0 h-full w-full" />;
}

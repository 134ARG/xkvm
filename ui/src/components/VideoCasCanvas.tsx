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
precision highp float;

uniform sampler2D uTexture;
uniform vec2 uInputSize;
uniform vec2 uOutputSize;
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

vec3 loadTexel(vec2 pixel) {
  vec2 clampedPixel = clamp(pixel, vec2(0.0), uInputSize - vec2(1.0));
  return texture2D(uTexture, (clampedPixel + vec2(0.5)) / uInputSize).rgb;
}

float casWeight(float mn, float mx, float peak) {
  float amp = clamp(min(mn, 1.0 - mx) / max(mx, 0.001), 0.0, 1.0);
  return sqrt(amp) * peak;
}

void main() {
  // Port of ffx_cas.h CasFilter() scaling path with default fast green-channel weights.
  vec2 outputPixel = floor(vec2(gl_FragCoord.x, uOutputSize.y - gl_FragCoord.y));
  vec2 scale = uInputSize / uOutputSize;
  vec2 sourcePosition = outputPixel * scale + 0.5 * scale - 0.5;
  vec2 basePixel = floor(sourcePosition);
  vec2 phase = sourcePosition - basePixel;

  vec3 b = loadTexel(basePixel + vec2( 0.0, -1.0));
  vec3 c = loadTexel(basePixel + vec2( 1.0, -1.0));
  vec3 e = loadTexel(basePixel + vec2(-1.0,  0.0));
  vec3 f = loadTexel(basePixel);
  vec3 g = loadTexel(basePixel + vec2( 1.0,  0.0));
  vec3 h = loadTexel(basePixel + vec2( 2.0,  0.0));
  vec3 i = loadTexel(basePixel + vec2(-1.0,  1.0));
  vec3 j = loadTexel(basePixel + vec2( 0.0,  1.0));
  vec3 k = loadTexel(basePixel + vec2( 1.0,  1.0));
  vec3 l = loadTexel(basePixel + vec2( 2.0,  1.0));
  vec3 n = loadTexel(basePixel + vec2( 0.0,  2.0));
  vec3 o = loadTexel(basePixel + vec2( 1.0,  2.0));

  float mnf = min(min(min(b.g, e.g), min(f.g, g.g)), j.g);
  float mxf = max(max(max(b.g, e.g), max(f.g, g.g)), j.g);
  float mng = min(min(min(c.g, f.g), min(g.g, h.g)), k.g);
  float mxg = max(max(max(c.g, f.g), max(g.g, h.g)), k.g);
  float mnj = min(min(min(f.g, i.g), min(j.g, k.g)), n.g);
  float mxj = max(max(max(f.g, i.g), max(j.g, k.g)), n.g);
  float mnk = min(min(min(g.g, j.g), min(k.g, l.g)), o.g);
  float mxk = max(max(max(g.g, j.g), max(k.g, l.g)), o.g);

  float peak = -1.0 / mix(8.0, 5.0, clamp(uSharpness, 0.0, 1.0));
  float wf = casWeight(mnf, mxf, peak);
  float wg = casWeight(mng, mxg, peak);
  float wj = casWeight(mnj, mxj, peak);
  float wk = casWeight(mnk, mxk, peak);

  float s = (1.0 - phase.x) * (1.0 - phase.y);
  float t = phase.x * (1.0 - phase.y);
  float u = (1.0 - phase.x) * phase.y;
  float v = phase.x * phase.y;

  float thin = 1.0 / 32.0;
  s /= thin + (mxf - mnf);
  t /= thin + (mxg - mng);
  u /= thin + (mxj - mnj);
  v /= thin + (mxk - mnk);

  float qbe = wf * s;
  float qch = wg * t;
  float qf = wg * t + wj * u + s;
  float qg = wf * s + wk * v + t;
  float qj = wf * s + wk * v + u;
  float qk = wg * t + wj * u + v;
  float qin = wj * u;
  float qlo = wk * v;
  float weightSum = 2.0 * qbe + 2.0 * qch + 2.0 * qin + 2.0 * qlo + qf + qg + qj + qk;

  vec3 color =
    (b * qbe + e * qbe + c * qch + h * qch + i * qin + n * qin +
     l * qlo + o * qlo + f * qf + g * qg + j * qj + k * qk) / weightSum;

  gl_FragColor = vec4(adjustColor(clamp(color, 0.0, 1.0)), 1.0);
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
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);

    const textureUniform = gl.getUniformLocation(program, "uTexture");
    const inputSizeUniform = gl.getUniformLocation(program, "uInputSize");
    const outputSizeUniform = gl.getUniformLocation(program, "uOutputSize");
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
      gl.uniform2f(inputSizeUniform, video.videoWidth, video.videoHeight);
      gl.uniform2f(outputSizeUniform, canvas.width, canvas.height);
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

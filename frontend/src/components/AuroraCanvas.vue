<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

// Single-scene WebGL aurora (raw GLSL, no three.js — offline friendly).
// Pauses when off-screen (IntersectionObserver), caps DPR on mobile, and
// degrades to a static CSS gradient under prefers-reduced-motion / no WebGL.
const props = withDefaults(defineProps<{ intensity?: number }>(), { intensity: 1 })

const canvas = ref<HTMLCanvasElement | null>(null)
const fallback = ref(false)

let gl: WebGLRenderingContext | null = null
let program: WebGLProgram | null = null
let raf = 0
let running = false
let start = 0
let io: IntersectionObserver | null = null
let uTime: WebGLUniformLocation | null = null
let uRes: WebGLUniformLocation | null = null
let uIntensity: WebGLUniformLocation | null = null

const reduce = () =>
  window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches

const VERT = `
attribute vec2 p;
void main(){ gl_Position = vec4(p, 0.0, 1.0); }
`

// Layered flow aurora in cyan→violet→magenta, soft vignette, subtle grain.
const FRAG = `
precision highp float;
uniform float u_time;
uniform vec2 u_res;
uniform float u_intensity;

float hash(vec2 p){ return fract(sin(dot(p, vec2(127.1,311.7)))*43758.5453); }
float noise(vec2 p){
  vec2 i=floor(p), f=fract(p);
  float a=hash(i), b=hash(i+vec2(1.,0.)), c=hash(i+vec2(0.,1.)), d=hash(i+vec2(1.,1.));
  vec2 u=f*f*(3.-2.*f);
  return mix(a,b,u.x)+(c-a)*u.y*(1.-u.x)+(d-b)*u.x*u.y;
}
float fbm(vec2 p){
  float v=0.0, a=0.5;
  for(int i=0;i<5;i++){ v+=a*noise(p); p*=2.03; a*=0.5; }
  return v;
}
void main(){
  vec2 uv = gl_FragCoord.xy / u_res.xy;
  vec2 q = uv;
  q.x *= u_res.x / u_res.y;
  float t = u_time * 0.05;

  // drifting flow field
  float f = fbm(q*1.7 + vec2(t, t*0.6));
  float f2 = fbm(q*2.6 - vec2(t*0.8, t*0.3) + f);
  float band = fbm(vec2(q.x*1.2 + f2*0.6, q.y*2.2 - t));

  vec3 cyan    = vec3(0.22, 0.91, 1.0);
  vec3 violet  = vec3(0.545, 0.36, 1.0);
  vec3 magenta = vec3(1.0, 0.24, 0.94);

  vec3 col = mix(cyan, violet, smoothstep(0.2, 0.8, f2));
  col = mix(col, magenta, smoothstep(0.45, 1.0, band));

  // aurora ribbons concentrated toward upper area, fading down
  float ribbon = smoothstep(0.35, 0.95, f2) * (1.0 - uv.y*0.6);
  float glow = pow(ribbon, 1.6) * 0.9;

  vec3 finalCol = col * glow * u_intensity;

  // vignette toward the void
  float vig = smoothstep(1.25, 0.2, length(uv-0.5));
  finalCol *= vig;

  // grain
  float g = (hash(gl_FragCoord.xy + u_time) - 0.5) * 0.03;
  finalCol += g;

  gl_FragColor = vec4(finalCol, 1.0);
}
`

function compile(glc: WebGLRenderingContext, type: number, src: string) {
  const s = glc.createShader(type)!
  glc.shaderSource(s, src)
  glc.compileShader(s)
  if (!glc.getShaderParameter(s, glc.COMPILE_STATUS)) {
    glc.deleteShader(s)
    return null
  }
  return s
}

function resize() {
  if (!gl || !canvas.value) return
  const dpr = Math.min(window.devicePixelRatio || 1, window.innerWidth < 768 ? 1 : 1.5)
  const w = Math.floor(canvas.value.clientWidth * dpr)
  const h = Math.floor(canvas.value.clientHeight * dpr)
  if (canvas.value.width !== w || canvas.value.height !== h) {
    canvas.value.width = w
    canvas.value.height = h
    gl.viewport(0, 0, w, h)
  }
}

function frame(now: number) {
  if (!running || !gl || !program) return
  if (!start) start = now
  resize()
  gl.uniform1f(uTime, (now - start) / 1000)
  gl.uniform2f(uRes, canvas.value!.width, canvas.value!.height)
  gl.uniform1f(uIntensity, props.intensity)
  gl.drawArrays(gl.TRIANGLES, 0, 6)
  raf = requestAnimationFrame(frame)
}

function play() {
  if (running || fallback.value) return
  running = true
  raf = requestAnimationFrame(frame)
}
function pause() {
  running = false
  if (raf) cancelAnimationFrame(raf)
  raf = 0
}

onMounted(() => {
  if (reduce() || !canvas.value) {
    fallback.value = true
    return
  }
  gl = canvas.value.getContext('webgl', { antialias: false, alpha: false, powerPreference: 'low-power' })
  if (!gl) {
    fallback.value = true
    return
  }
  const vs = compile(gl, gl.VERTEX_SHADER, VERT)
  const fs = compile(gl, gl.FRAGMENT_SHADER, FRAG)
  if (!vs || !fs) {
    fallback.value = true
    return
  }
  program = gl.createProgram()!
  gl.attachShader(program, vs)
  gl.attachShader(program, fs)
  gl.linkProgram(program)
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    fallback.value = true
    return
  }
  gl.useProgram(program)

  const buf = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, buf)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW)
  const loc = gl.getAttribLocation(program, 'p')
  gl.enableVertexAttribArray(loc)
  gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0)

  uTime = gl.getUniformLocation(program, 'u_time')
  uRes = gl.getUniformLocation(program, 'u_res')
  uIntensity = gl.getUniformLocation(program, 'u_intensity')

  resize()
  io = new IntersectionObserver(
    (entries) => {
      if (entries[0]?.isIntersecting) play()
      else pause()
    },
    { threshold: 0.01 }
  )
  io.observe(canvas.value)
  window.addEventListener('resize', resize, { passive: true })
})

onBeforeUnmount(() => {
  pause()
  io?.disconnect()
  window.removeEventListener('resize', resize)
})
</script>

<template>
  <div class="absolute inset-0 overflow-hidden">
    <canvas v-show="!fallback" ref="canvas" class="h-full w-full"></canvas>
    <!-- static gradient fallback (reduced-motion / no WebGL) -->
    <div
      v-if="fallback"
      class="h-full w-full"
      style="
        background:
          radial-gradient(60% 50% at 30% 25%, rgba(56, 232, 255, 0.22), transparent 60%),
          radial-gradient(55% 45% at 75% 30%, rgba(139, 92, 255, 0.18), transparent 60%),
          radial-gradient(60% 60% at 60% 80%, rgba(255, 61, 240, 0.16), transparent 60%);
      "
    ></div>
  </div>
</template>

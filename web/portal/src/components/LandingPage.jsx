import React, { useCallback, useEffect } from 'react'
import Navbar from './Navbar'
import CinematicHero from './cinematic/CinematicHero'
import StoryBeats from './cinematic/StoryBeats'
import ArchitectureDiagram from './ArchitectureDiagram'
import FeatureCards from './FeatureCards'
import OnboardingForm from './OnboardingForm'
import Footer from './Footer'

export default function LandingPage() {
  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'auto' })
  }, [])

  const scrollToForm = useCallback(() => {
    const el = document.getElementById('onboard')
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [])

  return (
    <div className="landing-root">
      <Navbar />
      <CinematicHero onScrollToForm={scrollToForm} />
      <StoryBeats />
      <ArchitectureDiagram />
      <FeatureCards />
      <OnboardingForm />
      <Footer />
    </div>
  )
}

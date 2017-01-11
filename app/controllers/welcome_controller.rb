class WelcomeController < ApplicationController

  def index
    render json: KyouenPuzzle.recent
  end
end

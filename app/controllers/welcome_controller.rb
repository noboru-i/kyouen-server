class WelcomeController < ApplicationController
  def index
    kyouen = Struct.new("Dog", :name, :age)
    kyouens = kyouen.new('welcome kyouen server', 98)
    render json: kyouens
  end
end

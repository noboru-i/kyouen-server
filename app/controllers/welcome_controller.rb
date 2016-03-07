class WelcomeController < ApplicationController
  def index
    kyouen = Struct.new("Dog", :name, :age)
    kyouens = kyouen.new('hoge', 7)
    render json: kyouens
  end
end
